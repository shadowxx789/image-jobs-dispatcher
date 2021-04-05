## Stub of "worker.cloud.net" service

### DESCRIPTION
This is stub which provide API for submitting job

Lang version: Go 1.15<br>
Linter: GolangCI 1.35.2 with list of rules in .golangci.yaml<br>
Base docker images:

    - Building: https://github.com/theshamuel/baseimg-go-build
    - Application: https://github.com/theshamuel/baseimg-go-app

#### API v1

1. Put blob `POST: /api/v1/job`
    - Request: 
        - <pre>{
            "encoding": "base64",
            "MD5":"[md5 hash of image file]",
            "content": "[encoded image file by algorithm from encoding field]'
          }
          </pre>
        - Ex: `curl --request POST \
          --url http://HOST:8080/api/v1/job \
          --header 'content-type: application/json' \
          --data '{
          "encoding": "base64",
          "MD5":"<md5 hash of image file>",
          "content": "<encoded image file by an algorithm from encoding field>"
          }'`

    - Response:
       - JSON:
         <pre>{
           "img_id":"[image id for getting blob from object store directly or via CDN]"
         }</pre>

1. Get blob data `GET: /api/v1/job/{id}/status`
    - Request: No Body
        - Ex: `curl --request GET \
          --url http://HOST:8081/api/v1/blob/1`
    - Response:
        - JSON:
          <pre>{
            "status":"one item from of the next enumeration [RUNNING | SUCCESS | FAILED]"
          }</pre>

### Improvements for using in a real pipeline as contract/smoke tests

    2. Add functionality to work with "worker.blob.net" for upload/get binary of images by chunks for making
    using network connection more balanced