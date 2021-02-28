package userv1

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/jmoiron/sqlx"
	"github.com/soldatov-s/go-garage-auth/models"
	"github.com/soldatov-s/go-garage/crypto/sha256"
	"github.com/soldatov-s/go-garage/providers/db"
	"github.com/soldatov-s/go-garage/types"
	"github.com/soldatov-s/go-garage/utils"
	"github.com/soldatov-s/go-garage/utils/email"
	"github.com/soldatov-s/go-garage/utils/phone"
	"github.com/soldatov-s/go-garage/x/sql"
)

type Partitions struct {
	TableName string `db:"table_name"`
}

const (
	checkInterval = 100 * time.Millisecond
)

func (u *UserV1) createUserPartitions(currentID int64) {
	var err error

	if u.db.Conn == nil {
		u.log.Err(db.ErrDBConnNotEstablished)
		return
	}

	if u.lastID == 0 {
		partitions := []Partitions{}
		if err = u.db.Conn.Select(&partitions,
			`SELECT "table_name" FROM information_schema.tables WHERE table_name LIKE 'user_' || '%' 
				AND table_schema = 'production'`); err != nil {
			u.log.Err(err)
			return
		}
		var partitionNumber int64
		for _, partition := range partitions {
			strs := strings.Split(partition.TableName, "_")
			if len(strs) != 3 {
				continue
			}

			partitionNumber, err = strconv.ParseInt(strs[len(strs)-1], 10, 64)
			if err != nil {
				u.log.Err(err).Msg("failed to convert partition number to string")
				return
			}

			u.log.Debug().Msgf("partition number from information_schema %d", partitionNumber)

			if partitionNumber > u.lastID {
				u.lastID = partitionNumber
			}
		}
	}

	// Mutex protects against duplication
	defer func() {
		if err1 := u.mu.Unlock(); err1 != nil {
			u.log.Err(err1).Msg("failed to unlock mutex")
		}
	}()

	if err = u.mu.Lock(); err != nil {
		return
	}

	if ((u.lastID - currentID) * 100 / u.lastID) <= 10 {
		partitionName := fmt.Sprintf("user_%d_%d", u.lastID+1, u.lastID+100000)
		// nolint G201: SQL string formatting (gosec)
		sqlRequest := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS production."%s" PARTITION OF production."user" FOR
		VALUES FROM (%d) TO (%d);
		CREATE INDEX IF NOT EXISTS %s_user_id ON production."%s" (user_id);`, partitionName,
			u.lastID+1, u.lastID+100001,
			partitionName, partitionName)
		_, err = u.db.Conn.Exec(u.db.Conn.Rebind(sqlRequest))
		if err != nil {
			u.log.Err(err).Msgf("failed to create user partition %d_%d", u.lastID+1, u.lastID+100000)
			return
		}
		u.lastID += 100000
		u.log.Debug().Msgf("created a new partition %s", partitionName)
	}
}

func (u *UserV1) createUser(c *models.NewCredentials) (data *models.User, err error) {
	// Normalize email
	normalEmail, err := email.Normilize(c.Email)
	if err != nil {
		return
	}

	// Normalize phone
	normalPhone, err := phone.Normilize(c.Phone)
	if err != nil {
		return
	}

	passwordHash, err := sha256.HashAndSalt(c.Password)
	if err != nil {
		return
	}

	data = &models.User{
		Hash:   passwordHash,
		Login:  c.Login,
		Email:  normalEmail,
		Phone:  normalPhone,
		Role:   c.Role,
		Status: c.Status,
	}

	data.CreateTimestamp()

	if u.db.Conn == nil {
		return nil, db.ErrDBConnNotEstablished
	}

	if u.createUserStmt == nil {
		u.createUserStmt, err = u.db.Conn.PrepareNamed(
			u.db.Conn.Rebind(utils.JoinStrings(" ", "INSERT INTO production.user", "("+strings.Join(data.SQLParamsRequest(), ", ")+")",
				"VALUES", "("+":"+strings.Join(data.SQLParamsRequest(), ", :")+") returning *")))
		if err != nil {
			return
		}
	}

	err = u.createUserStmt.Get(data, data)
	if err != nil {
		return nil, err
	}

	if data.ID == 0 {
		return nil, ErrLoginOrEmailIsOccupied
	}

	if err == nil {
		go u.createUserPartitions(data.ID)
	}

	return data, nil
}

func (u *UserV1) GetUserDataByID(id int64) (*models.User, error) {
	data := &models.User{}

	if err := sql.SelectByID(u.db.Conn, "production.user", id, &data); err != nil {
		return nil, err
	}

	return data, nil
}

func (u *UserV1) GetUserDataByCreds(c *models.Credentials) (data *models.User, err error) {
	data = &models.User{}

	if u.db.Conn == nil {
		return nil, db.ErrDBConnNotEstablished
	}

	// Get user from DB
	if c.Login != "" {
		err = u.db.Conn.Get(data, "select * from production.loginFastSearch($1)", c.Login)
	} else if c.Phone != "" {
		normolizedPhone, err1 := phone.Normilize(c.Phone)
		if err1 != nil {
			return nil, err1
		}

		err = u.db.Conn.Get(data, "select * from production.phonelFastSearch($1)", normolizedPhone)
	} else if c.Email != "" {
		// Normalize email
		normolizedEmail, err1 := email.Normilize(c.Email)
		if err1 != nil {
			return nil, err1
		}

		err = u.db.Conn.Get(data, "select * from production.emailFastSearch($1)", normolizedEmail)
	}

	if err != nil {
		return nil, err
	}

	// Check password
	err = sha256.ComparePasswords(data.Hash, c.Password)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (u *UserV1) checkSearchParameter(key string) bool {
	if key == "user_id" {
		return true
	}
	for _, v := range (&models.User{}).SQLParamsRequest() {
		if key == v {
			return true
		}
	}

	return false
}

func (u *UserV1) getUserDataByUserData(req *ArrayOfMapInterface) (data *ArrayOfUserData, err error) {
	fullQuery := "("
	queryMap := make(map[string]interface{})

	for i, item := range *req {
		var query []string

		for key, field := range item {
			if !u.checkSearchParameter(key) {
				return nil, ErrKeyDoNotMatch
			}

			if key == "user_email" {
				// Normalize email
				var (
					value string
					ok    bool
				)
				if value, ok = field.(string); !ok {
					return nil, ErrFailedTypeCast
				}
				normolizedEmail, err1 := email.Normilize(value)
				if err1 != nil {
					return nil, err1
				}

				queryMap[key+strconv.Itoa(i)] = normolizedEmail
			} else if key == "user_phone" {
				// Normalize phone
				var (
					value string
					ok    bool
				)
				if value, ok = field.(string); !ok {
					return nil, ErrFailedTypeCast
				}
				normolizedPhone, err1 := phone.Normilize(value)
				if err1 != nil {
					return nil, err1
				}

				queryMap[key+strconv.Itoa(i)] = normolizedPhone
			} else {
				queryMap[key+strconv.Itoa(i)] = field
			}

			if field == nil {
				query = append(query, key+" is null")
				continue
			}

			if key == "user_meta" {
				queryMap[key+strconv.Itoa(i)] = field.(types.NullMeta)
			}

			query = append(query, key+"=:"+key+strconv.Itoa(i))
		}

		fullQuery = fullQuery + strings.Join(query, " and ") + ") or ("
	}

	fullQuery = strings.TrimSuffix(fullQuery, " or (")

	if u.db.Conn == nil {
		return nil, db.ErrDBConnNotEstablished
	}

	var rows *sqlx.Rows
	if len(*req) > 0 {
		rows, err = u.db.Conn.NamedQuery("select * from production.user where ("+fullQuery+")", queryMap)
	} else {
		rows, err = u.db.Conn.Queryx("select * from production.user")
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	data = &ArrayOfUserData{}

	for rows.Next() {
		var item models.User

		err = rows.StructScan(&item)
		if err != nil {
			return nil, err
		}

		*data = append(*data, item)
	}

	return data, err
}

func (u *UserV1) mergeUserData(oldData *models.User, patch *[]byte) (newData *models.User, err error) {
	id := oldData.ID

	original, err := json.Marshal(oldData)
	if err != nil {
		return
	}

	merged, err := jsonpatch.MergePatch(original, *patch)
	if err != nil {
		return
	}

	// Save hash
	Hash := oldData.Hash
	ActivationHash := oldData.ActivationHash

	err = json.Unmarshal(merged, &newData)
	if err != nil {
		return
	}

	// Normalize email
	normalEmail, err := email.Normilize(newData.Email)
	if err != nil {
		return
	}
	newData.Email = normalEmail

	// Normalize phone
	normalPhone, err := phone.Normilize(newData.Phone)
	if err != nil {
		return
	}
	newData.Phone = normalPhone

	// Protect ID from changes
	newData.ID = id

	if newData.Hash == "" {
		newData.Hash = Hash
	}

	if !newData.ActivationHash.Valid {
		newData.ActivationHash = ActivationHash
	}

	err = newData.Validate()
	if err != nil {
		return nil, err
	}

	newData.UpdatedAt.SetNow()

	return newData, nil
}

type updateUserData struct {
	models.User
	NewLogin string `json:"new_login" db:"new_login"`
	NewMail  string `json:"new_email" db:"new_email"`
	NewPhone string `json:"new_phone" db:"new_phone"`
}

func (u *UserV1) updateUserByID(id int64, patch *[]byte) (writeData *models.User, err error) {
	data, err := u.GetUserDataByID(id)
	if err != nil {
		return
	}

	// Save old login/email/phone
	oldLogin := data.Login
	oldMail := data.Email
	oldPhone := data.Phone

	writeData, err = u.mergeUserData(data, patch)
	if err != nil {
		return
	}

	query := make([]string, 0, len(data.SQLParamsRequest()))
	for _, param := range data.SQLParamsRequest() {
		query = append(query, param+"=:"+param)
	}

	writeUpdateData := &updateUserData{}
	writeUpdateData.User = *writeData

	// Check that new mail/login/phone not equal old mail/login
	writeUpdateData.NewLogin = ""
	writeUpdateData.NewMail = ""
	writeUpdateData.NewPhone = ""
	if oldLogin != writeData.Login {
		writeUpdateData.NewLogin = writeData.Login
	}

	if oldMail != writeData.Email {
		writeUpdateData.NewMail = writeData.Email
	}

	if oldPhone != writeData.Phone {
		writeUpdateData.NewPhone = writeData.Phone
	}

	if u.db.Conn == nil {
		return nil, db.ErrDBConnNotEstablished
	}

	result, err := u.db.Conn.NamedExec(
		u.db.Conn.Rebind(utils.JoinStrings(" ", "UPDATE production.user SET", strings.Join(query, ", "),
			"WHERE NOT EXISTS (SELECT * FROM production.user WHERE user_email = :new_mail) AND user_id=:user_id")),
		writeUpdateData)
	if err != nil {
		return nil, err
	}

	countRow, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}

	if countRow == 0 {
		return nil, ErrLoginOrEmailIsOccupied
	}

	return writeData, err
}

func (u *UserV1) softDeleteUserByID(id int64) (err error) {
	data, err := u.GetUserDataByID(id)
	if err != nil {
		return
	}

	if data.DeletedAt.Valid {
		return nil
	}

	data.DeletedAt.Timestamp()

	query := make([]string, 0, len(data.SQLParamsRequest()))
	for _, param := range data.SQLParamsRequest() {
		query = append(query, param+"=:"+param)
	}

	if u.db.Conn == nil {
		return db.ErrDBConnNotEstablished
	}

	_, err = u.db.Conn.NamedExec(
		u.db.Conn.Rebind(utils.JoinStrings(" ", "UPDATE production.user SET", strings.Join(query, ", "), "WHERE user_id=:user_id")),
		data)

	return
}

func (u *UserV1) hardDeleteUserByID(id int64) (err error) {
	return sql.HardDeleteByID(u.db.Conn, "production.user", id)
}

func (u *UserV1) updateUserCredsByID(id int64, c *models.UpdateCredentials) (data *models.User, err error) {
	data, err = u.GetUserDataByID(id)
	if err != nil {
		return nil, err
	}

	// Check password
	if c.OldPassword != "" {
		err = sha256.ComparePasswords(data.Hash, c.OldPassword)
		if err != nil {
			return nil, err
		}
	}

	// Update password
	if c.Password != "" {
		// Checking that new password is not same as old password
		err = sha256.ComparePasswords(data.Hash, c.Password)
		if err != sha256.ErrMismatchedHashAndPassword {
			if err == nil {
				return nil, ErrNewPasswordIsSameAsOld
			}
			return nil, err
		}

		passwordHash, err1 := sha256.HashAndSalt(c.Password)
		if err1 != nil {
			return nil, err1
		}

		data.Hash = passwordHash
	}

	data.UpdatedAt.Timestamp()

	// We can not to use updateUserByID, because the json Data not include all hash-fields
	query := make([]string, 0, len(data.SQLParamsRequest()))
	for _, param := range data.SQLParamsRequest() {
		query = append(query, param+"=:"+param)
	}

	if u.db.Conn == nil {
		return nil, db.ErrDBConnNotEstablished
	}

	_, err = u.db.Conn.NamedExec(
		u.db.Conn.Rebind(utils.JoinStrings(" ", "UPDATE production.user SET", strings.Join(query, ", "), "WHERE user_id=:user_id")),
		data)
	if err != nil {
		return nil, err
	}

	return data, nil
}
