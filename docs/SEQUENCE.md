# Регистрация пользователя

```mermaid
sequenceDiagram
    actor Guest
    participant Gophermart
    Guest->>Gophermart:POST /api/user/register
    activate Guest
    activate Gophermart
    alt request body is invalid
        Gophermart-->>Guest:400 Bad Request
    else username is already taken
        Gophermart-->>Guest:409 Conflict
    else internal error
        Gophermart-->>Guest:500 Internal Server Error
    else else
        Gophermart-->>Guest:200 OK (Authorization: token)
    end
    deactivate Guest
    deactivate Gophermart
```

# Аутентификация пользователя

```mermaid
sequenceDiagram
    actor Guest
    participant Gophermart
    Guest->>Gophermart:POST /api/user/login
    activate Guest
    activate Gophermart
    alt request body is invalid
        Gophermart-->>Guest:400 Bad Request
    else incorrect username/password
        Gophermart-->>Guest:401 Unauthorized
    else internal error
        Gophermart-->>Guest:500 Internal Server Error
    else else
        Gophermart-->>Guest:200 OK (Authorization: token)
    end
    deactivate Guest
    deactivate Gophermart
```

# Загрузка номера заказа

```mermaid
sequenceDiagram
    actor User
    participant Gophermart
    User->>Gophermart:POST /api/user/orders (order_id)
    activate User
    activate Gophermart
    alt request body is invalid
        Gophermart-->>User:400 Bad Request
    else invalid authorization token
        Gophermart-->>User:401 Unauthorized
    else order_id is already submitted by another user
        Gophermart-->>User:409 Conflict
    else order_id is invalid
        Gophermart-->>User:422 Unprocessable Content
    else internal error
        Gophermart-->>User:500 Internal Server Error
    else order_id is already submitted by this user
        Gophermart-->>User:200 OK
    else order_id was not submitted before
        Gophermart-->>User:202 Accepted
    end
    deactivate User
    deactivate Gophermart
```

# Получение списка загруженных номеров заказов

```mermaid
sequenceDiagram
    actor User
    participant Gophermart
    User->>Gophermart:GET /api/user/orders
    activate User
    activate Gophermart
    alt invalid authorization token
        Gophermart-->>User:401 Unauthorized
    else internal error
        Gophermart-->>User:500 Internal Server Error
    else no submitted orders
        Gophermart-->>User:204 No Content
    else else
        Gophermart-->>User:200 OK [(number, status, accrual?, uploaded_at), ...]
        Note over User,Gophermart: accrual can be absent if order is not processed
    end
    deactivate User
    deactivate Gophermart
```

# Получение текущего баланса пользователя

```mermaid
sequenceDiagram
    actor User
    participant Gophermart
    User->>Gophermart:GET /api/user/balance
    activate User
    activate Gophermart
    alt invalid authorization token
        Gophermart-->>User:401 Unauthorized
    else internal error
        Gophermart-->>User:500 Internal Server Error
    else else
        Gophermart-->>User:200 OK (current, withdrawn)
    end
    deactivate User
    deactivate Gophermart
```

# Запрос на списание средств

```mermaid
sequenceDiagram
    actor User
    participant Gophermart
    User->>Gophermart:POST /api/user/balance/withdraw (order, sum)
    activate User
    activate Gophermart
    alt invalid authorization token
        Gophermart-->>User:401 Unauthorized
    else not enough points
        Gophermart-->>User:402 Payment Required
    else invalid order_id
        Gophermart-->>User:422 Unprocessable Content
    else internal error
        Gophermart-->>User:500 Internal Server Error
    else else
        Gophermart-->>User:200 OK
    end
    deactivate User
    deactivate Gophermart
```

# Получение информации о выводе средств

```mermaid
sequenceDiagram
    actor User
    participant Gophermart
    User->>Gophermart:GET /api/user/withdrawals
    activate User
    activate Gophermart
    alt invalid authorization token
        Gophermart-->>User:401 Unauthorized
    else internal error
        Gophermart-->>User:500 Internal Server Error
    else no withdrawals
        Gophermart-->>User:204 No Content
    else else
        Gophermart-->>User:200 OK [(order, sum, processed_at), ...]
    end
    deactivate User
    deactivate Gophermart
```
