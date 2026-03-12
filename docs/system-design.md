## A. Functional Requirements:
1. The system must support user Registration and login using JWT Token
2. CRUD Tasks and only allow access to Authenticated User.
3. Unit Testing for Service layer and Repository Layer.


## B. Non-functional requirements:
1. Security, using Authentication and authorization to make sure the APIs is secure.
2. Logging, we need logging because:
    - To call attention to errors or add context before crashing, 
    - Help us to tracing of program execution,
    - Capture potentially problematic state in warning
3. Error Handling is crucial here.
4. Documentation, will be helpful to let other devs understanding the system.

## C. API Layer:
### POST
| Endpoint Name      | Input                                                                           |
|--------------------|---------------------------------------------------------------------------------|
| /v1/users/register | name, email, passsword                                                          |
| /v1/user/login     | email, password                                                                 |
| /v1/tasks          | title, description, due_date                                                    |

### GET
| Endpoint Name        | Params                                             |
|----------------------|----------------------------------------------------|
| /v1/tasks?query      | page, example: /v1/tasks?page=1                    |
|                      | limit, example: /v1/tasks?limit=10                 |
|                      | page&limit, example: /v1/tasks?page=1&limit=10     |
|                      | status, example: /v1/tasks?status=done             |
|                      | search, example: /v1/tasks?search=backend          |
| /v1/tasks/{id}       | taskId, example: /v1/tasks/1                       |
| /v1/users/{id}       | userId, example: /v1/users/1                       |
| /v1/users            | N/A                                                |
| /v1/users/my-profile | JWT_auth_token                                     |

### PUT
| Endpoint Name      | Params                                                                    |
|--------------------|---------------------------------------------------------------------------|
| /v1/users/:id      | name, email                                                               |
| /v1/tasks/:id      | tasks, status                                                             |

### DELETE
| Endpoint Name      | Params                                                                    |
|--------------------|---------------------------------------------------------------------------|
| /v1/tasks/:id      | N/A                                                                       |


## D. Database Layer
| User Table |                                     |
|------------|-------------------------------------|
| id         | userId AS UUID PRIMARY KEY NOT NULL |
| name       | VARCHAR                             |
| email      | VARCHAR NOT NULL                    |
| password   | VARCHAR NOT NULL                    |
| createdAt  | Timestamp                           |
| updatedAt  | Timestamp                           |
| deletedAt  | Timestamp                           |

| Task Table      |                                                                             |
|-----------------|-----------------------------------------------------------------------------|
| id              | UUID PRIMARY KEY NOT NULL                                                   |
| user_id         | UUID REFERENCES users(id) ON DELETE SET NULL, -- Link to User Table         |
| title           | Text NOT NULL                                                               |
| description     | Text.                                                                       |
| status          | Text (todo, in_progress, done)                                              |
| due_date        | DATE NOT NULL -- must be valid Date                                         |
| created_at      | Timestamp                                                                   |
| updated_at      | Timestamp                                                                   |           
| deletedAt       | Timestamp                                                                   |
