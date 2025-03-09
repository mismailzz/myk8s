# myk8s

```
curl -X POST "http://localhost:8080/createPod" -d '{"name": "nginx", "image": "nginx:latest"}' -H "Content-Type: application/json"
```

```
curl -X GET "http://localhost:8080/listPods"
```

```
curl -X DELETE "http://localhost:8080/deletePod?id=<POD_ID>"
```
