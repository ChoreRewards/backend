# Backend

Provides a gRPC and optional HTTP backend for ChoreRewards.

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
