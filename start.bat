@echo on
start "gate" cmd /k "cd gate && go run main.go"
start "auth" cmd /k "cd auth && go run main.go"
timeout /T 2 /NOBREAK
start "group" cmd /k "cd group && go run main.go"
start "proxy" cmd /k "cd proxy && go run main.go"
start "thing" cmd /k "cd thing && go run main.go"
timeout /T 2 /NOBREAK
start "shadow" cmd /k "cd shadow && go run main.go"
timeout /T 2 /NOBREAK
start "route" cmd /k "cd route && go run main.go"
exit