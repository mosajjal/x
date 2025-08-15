# Create a Docker image from a binary

This program allows you to create an exported Docker image from a single statically linked binary. This works against multiple architectures.

## Usage

```
go run . <binary>
```

above will generate "image.tar" in the current directory, and puts the binary under `/app/` directory inside the container

## Multi-platform build

1. Build the images for both amd64 and arm64 architectures

```
go run . <amd64-binary>
mv image.tar amd64-image.tar

go run . <arm64-binary>
mv image.tar arm64-image.tar
```

2. import both images and tag them appropriately

```
docker load -i amd64-image.tar
docker tag <image-id> your-registry.com/my-app:latest-amd64

docker load -i arm64-image.tar
docker tag <image-id> your-registry.com/my-app:latest-arm64
```

3. push each image individually

```
docker push your-registry.com/my-app:latest-amd64
docker push your-registry.com/my-app:latest-arm64
```

3. create a manifest and amend it with the tags

```
docker manifest create your-registry.com/my-app:latest \
    your-registry.com/my-app:latest-amd64 \
    your-registry.com/my-app:latest-arm64
```

4. push the manifest to the registry

```
docker manifest push your-registry.com/my-app:latest
```
