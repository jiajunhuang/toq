BUILD=go build
FLAGS=
FMT=go fmt

all: fmt main worker

fmt:
	$(FMT) github.com/jiajunhuang/toq/...

main:
	$(BUILD) $(FLAGS) -o main main.go utils.go

worker:
	$(BUILD) $(FLAGS) -o worker worker.go utils.go

clean:
	rm -f main worker
