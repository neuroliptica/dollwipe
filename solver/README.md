# Manual solver backend
Ручной ввод капчи, адрес - `localhost:8080`

## Собрать

```bash
$ make build
```

## Routes

### client API
- `GET /get`
    - получить капчу: хэш и картинку в base64
- `POST /solve`
    - решить капчу по хэшу

### application API 
- `POST /post`
    - запостить капчу и получить хэш
- `GET /check`
    - проверить статус капчи по хэшу

