
run: www
	./www -addr=:8081

www: *.go *.html *.css blog/*.*
	go generate
	go build -i
