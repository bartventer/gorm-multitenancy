### API Usage

#### Create tenant

-   Parse the request body into a CreateTenantBody struct
-   Create the tenant in the database (public schema)
-   Create the schema for the tenant
-   Return the HTTP status code 201 and the tenant in the response body

##### Request

```bash
curl -X POST \
  http://example.com:8080/tenants \
  -H 'Content-Type: application/json' \
  -d '{
  "domainUrl": "tenant3.example.com"
}'
```

##### Response

```json
{
    "id": 3,
    "domainUrl": "tenant3.example.com"
}
```

#### Get tenant

-   Get the tenant from the database
-   Return the HTTP status code 200 and the tenant in the response body

##### Request

```bash
curl -X GET \
  http://example.com:8080/tenants/3
```

##### Response

```json
{
    "id": 3,
    "domainUrl": "tenant3.example.com"
}
```

#### Delete tenant

-   Get the tenant from the database
-   Delete the schema for the tenant
-   Delete the tenant from the database
-   Return the HTTP status code 204

##### Request

```bash
curl -X DELETE \
  http://example.com:8080/tenants/3
```

##### Response

```json

```

#### Get books

-   Get the tenant from the request host or header
-   Get all books for the tenant
-   Return the HTTP status code 200 and the books in the response body

##### Request

```bash
curl -X GET \
  http://example.com:8080/books \
  -H 'Host: tenant1.example.com'
```

##### Response

```json
[
    {
        "id": 1,
        "name": "tenant1 - Book 1"
    },
    {
        "id": 2,
        "name": "tenant1 - Book 2"
    }
]
```

#### Create book

-   Get the tenant from the request host or header
-   Parse the request body into a Book struct
-   Create the book for the tenant in the database
-   Return the HTTP status code 201 and the book in the response body

##### Request

```bash
curl -X POST \
  http://example.com:8080/books \
  -H 'Content-Type: application/json' \
  -H 'Host: tenant1.example.com' \
  -d '{
  "name": "tenant1 - Book 3"
}'
```

##### Response

```json
{
    "id": 3,
    "name": "tenant1 - Book 3"
}
```

#### Delete book

-   Get the tenant from the request host or header
-   Get the book from the database
-   Delete the book from the database
-   Return the HTTP status code 204

##### Request

```bash
curl -X DELETE \
  http://example.com:8080/books/3 \
  -H 'Host: tenant1.example.com'
```

##### Response

```json

```

#### Update book

-   Get the tenant from the request host or header
-   Get the book from the database
-   Parse the request body into a UpdateBookBody struct
-   Update the book in the database
-   Return the HTTP status code 200

##### Request

```bash
curl -X PUT \
  http://example.com:8080/books/2 \
  -H 'Content-Type: application/json' \
  -H 'Host: tenant1.example.com' \
  -d '{
  "name": "tenant1 - Book 2 - Updated"
}'
```

##### Response

```json

```
