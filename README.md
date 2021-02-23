# Example project with go-garage

This is a simple identity provider, which created on base go-garage.

## How starts
```bash
make docker-compose-up
```
Will be started:
* postgresql
* go-garage-auth service  

Service applies postgresql migrations by self.  
After starting you can find swaggers:
* http://localhost:9000/api/v1/swagger/index.html
* http://localhost:9100/api/v1/swagger/index.html  

Prometheus metrics http://localhost:9100/metrics  
Alive http://localhost:9100/health/alive  
Ready http://localhost:9100/health/ready  