package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/cschleiden/go-workflows/backend"
	"github.com/cschleiden/go-workflows/backend/mysql"

	"github.com/cschleiden/go-workflows/client"
	"github.com/cschleiden/go-workflows/worker"
	"github.com/cschleiden/go-workflows/workflow"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

func Workflow1(ctx workflow.Context, searchTerms []string) error {
	//@see https://github.com/github/drifter/blob/2ca7907175c899a786864f9e08d7dadb2a0b137f/run/orch/workflows/runjobfactory.go
	//@see https://github.com/github/drifter/blob/81f3cd335814eecb105c450a42540f96b78bdf0d/run/orch/workflows/runworkflow.go

	// Channel used to control max parallelism of jobs. Its capacity determines how many jobs can run in parallel.
	maxJobs := len(searchTerms)

	//Create channel to store output of goroutine below
	c := workflow.NewBufferedChannel[struct{}](maxJobs)

	completed := make(map[string]string)

	for _, term := range searchTerms {

		term := term // Why do we need to copy this in order to work?

		// Start sub-workflow for each job
		workflow.Go(ctx, func(ctx workflow.Context) {
			wr, err := workflow.CreateSubWorkflowInstance[string](ctx, workflow.SubWorkflowOptions{}, Workflow2, term).Get(ctx)

			if err != nil {
				panic("could not get sub workflow result")
			}

			completed[term] = wr

			// Send value
			c.Send(ctx, struct{}{})
		})

		// Receive all values from channel
		c.Receive(ctx)
	}

	for term, value := range completed {
		fmt.Printf("Search term: %s\nTweet: %s\n\n", term, value)
	}

	return nil
}

func Workflow2(ctx workflow.Context, term string) (string, error) {
	r1, err := workflow.ExecuteActivity[string](ctx, workflow.DefaultActivityOptions, Activity1, term).Get(ctx)

	if err != nil {
		panic("error getting activity 1 result")
	}

	return r1, nil
}

func Activity1(ctx context.Context, term string) (string, error) {

	// oauth2 configures a client that uses app credentials to keep a fresh token
	config := &clientcredentials.Config{
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		TokenURL:     "https://api.twitter.com/oauth2/token",
	}
	// http.Client will automatically authorize Requests
	httpClient := config.Client(oauth2.NoContext)

	// Twitter client
	client := twitter.NewClient(httpClient)

	search, _, err := client.Search.Tweets(&twitter.SearchTweetParams{
		Query: term,
	})

	if err != nil {
		panic("Can't get tweets")
	}

	return search.Statuses[0].Text, nil
}

func runWorker(ctx context.Context, mb backend.Backend) {
	w := worker.New(mb, nil)

	w.RegisterWorkflow(Workflow1)
	w.RegisterWorkflow(Workflow2) // Register both workflows!

	w.RegisterActivity(Activity1)

	if err := w.Start(ctx); err != nil {
		panic("could not start worker")
	}
}

func main() {
	ctx := context.Background()

	backend := mysql.NewMysqlBackend("localhost", 3306, "root", "root", "simple")

	go runWorker(ctx, backend)

	c := client.New(backend)

	// Dummy primes data for workflow
	searchTerms := []string{"Winter", "Snow", "Hockey", "Maple Syrup"}

	wf, err := c.CreateWorkflowInstance(ctx, client.WorkflowInstanceOptions{
		InstanceID: uuid.NewString(),
	}, Workflow1, searchTerms)

	if err != nil {
		panic("could not start workflow")
	}

	log.Println("Running workflow: ", wf.InstanceID, wf.ExecutionID)

	c2 := make(chan os.Signal, 1)
	signal.Notify(c2, os.Interrupt)
	<-c2
}
