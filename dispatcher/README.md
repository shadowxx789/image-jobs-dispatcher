## Image Jobs dispatcher

Service to accepting jobs to could image farm.

Lang version: Go 1.15<br>
Linter: GolangCI 1.35.2 with list of rules in .golangci.yaml<br>
Base docker images:

    - Building: https://github.com/theshamuel/baseimg-go-build
    - Application: https://github.com/theshamuel/baseimg-go-app

### API Description v1

1. Submit job `POST: /api/v1/job` `Headers: Content-Type: application/json, Authorization: Bearer <JWT>`
    - JWT payload structure:
       <pre>
        {
            "sub": "1234567890",
            "name": "John Doe",
            "iat": 1516239022,
            "tid": 1,
            "oid": 1,
            "aud": "com.company.jobservice",
            "azp": "1",
            "email": "customer@mail.com"
        }
        </pre>
    - Ex:
      eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJ0aWQiOjEsIm9pZCI6MSwiYXVkIjoiY29tLmNvbXBhbnkuam9ic2VydmljZSIsImF6cCI6IjEiLCJlbWFpbCI6ImN1c3RvbWVyQG1haWwuY29tIn0.CcTapGbWX0UEMovUwC8kAcWMUxmbOeO0qhsu-wqHQH0
    - Request:
        - JSON:
            <pre>
            {
                "encoding": "base64",
                "MD5":"[md5 hash]",
                "content""[base64 hash]":
            }
            </pre>
    - Response:
        - JSON:
          <pre>{
            "id":"1"
          }</pre>
1. Get jobs result
   job `GET: /api/v1/job/{id}`
    - Request: No Body
        - Ex: `curl --request GET \
          --url http://localhost:8081/api/v1/blob/1`
    - Response:
        <pre>
        {
            "id": "6",
            "tenant_id": 1,
            "client_id": 1,
            "payload_location": "/images/1" #Image location in object store or CDN
        }
        </pre>

1. Get job status `GET: /api/v1/job/{id}/status`
    - Request: No Body
        - Ex: `curl --request GET \
          --url http://HOST:8081/api/v1/blob/1`
    - Response:
        - JSON:
          <pre>{
            "status":"one item from of the next enumeration [RUNNING | SUCCESS | FAILED]"
          }</pre>
      For id = 1 status = RUNNING, id = 2 status = SUCCESS, id = 3 status FAILED
