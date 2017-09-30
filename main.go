package main

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"fmt"
	"strings"
	"os"
	"log"
	"github.com/mitchellh/cli"
	"sort"
	"bytes"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/hashicorp/terraform/terraform"
	"github.com/hashicorp/terraform/builtin/providers/aws"
	"github.com/hashicorp/terraform/config"
	"sync"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/aws/aws-sdk-go/service/kms"
)

func main() {
	app := "awsweeper"
	profile := ""
	if len(os.Args) > 1 {
		profile = os.Args[1]
	} else {
		fmt.Println("Profile is missing")
		os.Exit(1)
	}

	//log.SetFlags(0)
	//log.SetOutput(ioutil.Discard)

	c := &cli.CLI{
		Name: app,
		Version: "0.0.1",
		HelpFunc: BasicHelpFunc(app),
	}
	c.Args = os.Args[2:]

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Profile: profile,
	}))
	region := *sess.Config.Region

	p := initAwsProvider(profile, region)

	//f := flagSet("bla")

	client := &AWSClient{
		autoscalingconn: autoscaling.New(sess),
		ec2conn: ec2.New(sess),
		elbconn: elb.New(sess),
		r53conn: route53.New(sess),
		cfconn: cloudformation.New(sess),
		efsconn:  efs.New(sess),
		iamconn: iam.New(sess),
		kmsconn: kms.New(sess),
	}

	c.Commands = map[ string]cli.CommandFactory{
		"yaml": func() (cli.Command, error) {
			return &WipeByYamlConfig{
				client: client,
				provider: p,
				yamlConfig: map[string]B{},
			}, nil
		},
		"all": func() (cli.Command, error) {
			return &WipeAllCommand{
				autoscalingconn: autoscaling.New(sess),
				ec2conn: ec2.New(sess),
				elbconn: elb.New(sess),
				r53conn: route53.New(sess),
				cfconn: cloudformation.New(sess),
				efsconn:  efs.New(sess),
				iamconn: iam.New(sess),
				kmsconn: kms.New(sess),
				provider: p,
			}, nil
		},
		"output": func() (cli.Command, error) {
			return &WipeAllCommand{
				autoscalingconn: autoscaling.New(sess),
				ec2conn: ec2.New(sess),
				elbconn: elb.New(sess),
				r53conn: route53.New(sess),
				cfconn: cloudformation.New(sess),
				efsconn:  efs.New(sess),
				iamconn: iam.New(sess),
				kmsconn: kms.New(sess),
				provider: p,
				out: map[string]B{},
			}, nil
		},
	}

	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}

	os.Exit(exitStatus)
}

func initAwsProvider(profile string, region string) *terraform.ResourceProvider {
	p := aws.Provider()

	// list all schemas of resource types
	//for k, v := range aws.Provider().(*schema.Provider).ResourcesMap {
	//	fmt.Println(k)
	//	for k, v := range  v.Schema {
	//		fmt.Println("\t", k)
	//		fmt.Println("\t", v)
	//	}
	//}

	// list all resource types
	//for _, r := range p.Resources() {
	//	fmt.Println(r)
	//}

	cfg := map[string]interface{}{
		"region":     region,
		"profile":    profile,
	}

	rc, err := config.NewRawConfig(cfg)
	if err != nil {
		fmt.Printf("bad: %s\n", err)
		os.Exit(1)
	}
	conf := terraform.NewResourceConfig(rc)

	warns, errs := p.Validate(conf)
	if len(warns) > 0 {
		fmt.Printf("warnings: %s\n", warns)
	}
	if len(errs) > 0 {
		fmt.Printf("errors: %s\n", errs)
		os.Exit(1)
	}

	if err := p.Configure(conf); err != nil {
		fmt.Printf("err: %s\n", err)
		os.Exit(1)
	}

	return &p
}

/*
// flags adds the meta flags to the given FlagSet.
func flagSet(n string) *flag.FlagSet {
	f := flag.NewFlagSet(n, flag.ContinueOnError)
	f.String("var", "",  "variables")
	f.String("var", "",  "tag-value")

	// Create an io.Writer that writes to our Ui properly for errors.
	// This is kind of a hack, but it does the job. Basically: create
	// a pipe, use a scanner to break it into lines, and output each line
	// to the UI. Do this forever.
	errR, errW := io.Pipe()
	errScanner := bufio.NewScanner(errR)
	go func() {
		for errScanner.Scan() {
			m.Ui.Error(errScanner.Text())
		}
	}()
	f.SetOutput(errW)

	// Set the default Usage to empty
	f.Usage = func() {}

	return f
}
*/

func BasicHelpFunc(app string) cli.HelpFunc {
	return func(commands map[string]cli.CommandFactory) string {
		var buf bytes.Buffer
		buf.WriteString(fmt.Sprintf(
			"Usage: %s [--version] [--help] <profile> <command> [<args>]\n\n",
			app))
		buf.WriteString("Available commands are:\n")

		// Get the list of keys so we can sort them, and also get the maximum
		// key length so they can be aligned properly.
		keys := make([]string, 0, len(commands))
		maxKeyLen := 0
		for key := range commands {
			if len(key) > maxKeyLen {
				maxKeyLen = len(key)
			}

			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, key := range keys {
			commandFunc, ok := commands[key]
			if !ok {
				// This should never happen since we JUST built the list of
				// keys.
				panic("command not found: " + key)
			}

			command, err := commandFunc()
			if err != nil {
				log.Printf("[ERR] cli: Command '%s' failed to load: %s",
					key, err)
				continue
			}

			key = fmt.Sprintf("%s%s", key, strings.Repeat(" ", maxKeyLen - len(key)))
			buf.WriteString(fmt.Sprintf("    %s    %s\n", key, command.Synopsis()))
		}

		return buf.String()
	}
}

type Resource struct {
	id *string
	attrs *map[string]string
	tags *map[string]string
}

func (c *WipeCommand) deleteResources(rSet ResourceSet) {
	isDryRun := true
	numWorkerThreads := 10

	if len(rSet.Ids) == 0 {
		return
	}

	c.out[rSet.Type] = B{Ids: rSet.Ids}

	printType(rSet.Type, len(rSet.Ids))
	if len(rSet.Info) > 0 {
		fmt.Println(rSet.Info)
	}

	ii := &terraform.InstanceInfo{
		Type: rSet.Type,
	}

	d := &terraform.InstanceDiff{
		Destroy: true,
	}

	a := []*map[string]string{}
	if len(rSet.Attrs) > 0 {
		a = rSet.Attrs
	} else {
		for i := 0; i < len(rSet.Ids); i++ {
			a = append(a, &map[string]string{})
		}
	}

	ts := make([]*map[string]string, len(rSet.Ids))
	if len(rSet.Tags) > 0 {
		ts = rSet.Tags
	}
	chResources := make(chan *Resource, numWorkerThreads)

	var wg sync.WaitGroup
	wg.Add(len(rSet.Ids))

	for j := 1; j <= numWorkerThreads; j++ {
		go func() {
			for {
				res, more := <- chResources
				if more {
					printStat := fmt.Sprintf("\tId:\t%s", *res.id)
					if res.tags != nil {
						if len(*res.tags) > 0 {
							printStat += "\n\tTags:\t"
							for k, v := range *res.tags {
								printStat += fmt.Sprintf("[%s: %v] ", k, v)
							}
							printStat += "\n"
						}
					}
					fmt.Println(printStat)

					a := res.attrs
					(*a)["force_destroy"] = "true"

					s := &terraform.InstanceState{
						ID: *res.id,
						Attributes: *a,
					}

					st, err := (*c.provider).Refresh(ii, s)
					if err != nil{
						fmt.Println("err: ", err)
						st = s
						st.Attributes["force_destroy"] = "true"
					}
					if rSet.Type == "aws_iam_role_policy_attachment" {
						fmt.Println(st)
					}
					if rSet.Type == "aws_iam_role" {
						fmt.Println(st)
					}

					if !isDryRun {
						_, err := (*c.provider).Apply(ii, st, d)

						if err != nil {
							fmt.Printf("\t%s\n", err)
						}
					}
					wg.Done()
				} else {
					return
				}
			}
		}()
	}

	for i, id := range rSet.Ids {
		if id != nil {
			chResources <- &Resource{
				id: id,
				attrs: a[i],
				tags: ts[i],
			}
		}
	}
	close(chResources)

	wg.Wait()
	fmt.Println("---\n")
}

func printType(resourceType string, numberOfResources int) {
	fmt.Printf("\n---\nType: %s\nFound: %d\n\n", resourceType, numberOfResources)
}

func HasPrefix(s string, prefixes []string) bool {
	result := false
	for _, prefix := range prefixes{
		if strings.HasPrefix(s, prefix) {
			result = true
		}
	}
	return result
}
