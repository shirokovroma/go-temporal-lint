package testdata

import (
	"context"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

// HelloWorldWorkflow is the workflow definition.
func HelloWorldWorkflow(ctx workflow.Context, name string) (string, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var result string
	// The activity function HelloWorldActivity accepts 2 arguments, but 3 were passed
	err := workflow.ExecuteActivity(ctx, HelloWorldActivity, name, name).Get(ctx, &result)
	if err != nil {
		return "", err
	}

	return result, nil
}

// HelloWorldActivity is the activity definition.
func HelloWorldActivity(ctx context.Context, name string) (string, error) {
	return "Hello " + name + "!", nil
}

func main() {
	// Create the client object just once per process.
	c, err := client.NewClient(client.Options{})
	if err != nil {
		panic(err)
	}
	defer c.Close()

	// This worker hosts both HelloWorldWorkflow and HelloWorldActivity.
	w := worker.New(c, "hello-world-task-queue", worker.Options{})
	w.RegisterWorkflow(HelloWorldWorkflow)
	w.RegisterActivity(HelloWorldActivity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		panic(err)
	}
}

func dummyWorkflowCall() {
	// The activity function HelloWorldActivity accepts 2 arguments, but 3 were passed
	workflow.ExecuteActivity(nil, HelloWorldActivity, "name", 1)
	// In the function HelloWorldActivity, the type of argument 2 is string, but int was passed
	workflow.ExecuteActivity(nil, HelloWorldActivity, 1)
}
