OS=linux
ARCH=amd64

imagever:
			docker build -t theshamuel/worker-service-mock:1.0.0 blob-service-mock
			docker build -t theshamuel/blob-service-mock:1.0.0 worker-service-mock
			docker build -t theshamuel/image-jobs-dispatcher:1.0.0 dispatcher

imagedev:
			docker build -t theshamuel/worker-service-mock:1.0.0 --build-arg SKIP_TESTS=true blob-service-mock
			docker build -t theshamuel/blob-service-mock:1.0.0 --build-arg SKIP_TESTS=true worker-service-mock
			docker build -t theshamuel/image-jobs-dispatcher:1.0.0 --build-arg SKIP_TESTS=true dispatcher

deploy:
			docker-compose up -d

undeploy:
			docker-compose down