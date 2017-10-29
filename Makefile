
run: www
	./www -addr=:8081

www: *.go *.html
	go generate
	go build -i
