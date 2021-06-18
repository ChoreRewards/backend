# Backend

Provides a gRPC and optional HTTP backend for ChoreRewards.

# gRPC requests

## Pre-requisites

- Install [grpcurl](https://github.com/fullstorydev/grpcurl)

## List services

```
grpcurl -plaintext localhost:8080 list
```

## List RPC endpoints

Note: This assumes the api repository is cloned at the same location as this repository (`../api`), that the required dependencies have been install into the `.cache` directory, of the api repository, and that you're using an Apple device (Darwin).

```
grpcurl -protoset <(cd ../api; ../api/.cache/Darwin/x86_64/bin/buf image build -o -) -plaintext localhost:8080 list chorerewards.v1alpha1.ChoreRewardsService
```

## Create a user

```
grpcurl -protoset <(cd ../api; ../api/.cache/Darwin/x86_64/bin/buf image build -o -) -plaintext -d '{"user": { "username": "testUser2", "email": "user@example.com", "password": "password", "pin": 1234 } }' localhost:8080 chorerewards.v1alpha1.ChoreRewardsService/CreateUser
```

## Login

```
// With Password
grpcurl -protoset <(cd ../api; ../api/.cache/Darwin/x86_64/bin/buf image build -o -) -plaintext -d '{"username": "testUser", "password": "password" }' localhost:8080 chorerewards.v1alpha1.ChoreRewardsService/Login

// With Pin
grpcurl -protoset <(cd ../api; ../api/.cache/Darwin/x86_64/bin/buf image build -o -) -plaintext -d '{"username": "testUser", "pin": 1234 }' localhost:8080 chorerewards.v1alpha1.ChoreRewardsService/Login
```

# HTTP requests

## Create a user

```
curl -H "Content-Type: application/json" -X POST localhost:8443/v1alpha1/users -d '{"username": "testUser", "email": "user@example.com", "password": "password", "pin": 1234}'
```

## Login

```
// With Password
curl -H "Content-Type: application/json" -X POST localhost:8443/v1alpha1/login -d '{"username": "testUser", "password": "password"}

// With Pin
curl -H "Content-Type: application/json" -X POST localhost:8443/v1alpha1/login -d '{"username": "testUser", "pin": 1234}'
```
