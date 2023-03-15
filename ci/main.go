package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
)

func main() {
	if err := executeTask(context.Background()); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func executeTask(ctx context.Context) error {

	fmt.Println("ci process started")

	taskId := os.Args[1]
	if taskId == "" {
		return errors.New("argonaut build identifier is missing")
	}

	userRepoLoc := os.Args[2]
	if userRepoLoc == "" {
		return errors.New("user repo location missing")
	}

	fmt.Printf("taskId [%s] userRepoLoc [%s] \n", taskId, userRepoLoc)

	_, err := InitializeArgoClient()
	if err != nil {
		fmt.Printf("Argonaut client setup failed : [%v] \n", err)
		return err
	}

	fmt.Printf("Argonaut client setup complete! \n")

	switch {
	case strings.HasPrefix(taskId, "br-"):
		return build(ctx, strings.TrimPrefix(taskId, "br-"), userRepoLoc)
	default:
		return errors.New("unknown task type")
	}
}
