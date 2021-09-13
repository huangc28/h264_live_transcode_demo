# Spin up backend
run_local:
	echo 'running...'

	go mod tidy && go run . -alsologtostderr=true