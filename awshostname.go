package main

import (
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"os"
	"sort"
	"strconv"
	"strings"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [options] <hostname>\n", os.Args[0])
	flag.PrintDefaults()
}

func parseTags(filters []*ec2.Filter, input *string, spec *HostSpec) ([]*ec2.Filter, error) {
	if len(*input) == 0 {
		return filters, nil
	}
	tagPairs := strings.Split(*input, ",")
	for _, pair := range tagPairs {
		kv := strings.Split(pair, "=")
		if len(kv) != 2 || len(kv[0]) == 0 || len(kv[1]) == 0 {
			return filters, fmt.Errorf("Invalid tag spec: %s", pair)
		}
		v, err := replacePositional(kv[1], spec)
		if err != nil {
			return nil, err
		}
		filters = addFilter(filters, fmt.Sprintf("tag:%s", kv[0]), v)
	}
	return filters, nil
}

func replacePositional(s string, spec *HostSpec) (string, error) {
	if len(s) > 1 && s[0] == '#' {
		i, err := strconv.Atoi(s[1:])
		if err != nil {
			return s, err
		}
		if i < 0 || i >= len(spec.Names) {
			return s, fmt.Errorf("No %d'th component exists", i)
		}
		s = spec.Names[i]
	}
	return s, nil
}

func addFilter(filters []*ec2.Filter, name string, value string) []*ec2.Filter {
	filters = append(filters, &ec2.Filter{
		Name:   aws.String(name),
		Values: []*string{aws.String(value)},
	})
	return filters
}

type EC2Instances []*ec2.Instance

func (instances EC2Instances) Len() int {
	return len(instances)
}

func (instances EC2Instances) Swap(i, j int) {
	instances[i], instances[j] = instances[j], instances[i]
}

func (instances EC2Instances) Less(i, j int) bool {
	return (*instances[i].LaunchTime).Before(*instances[j].LaunchTime)
}

type HostSpec struct {
	Names []string
	Index int
}

func parseHostName(s *string) (*HostSpec, error) {
	var h HostSpec
	h.Index = -1
	h.Names = strings.Split(*s, ".")
	idx := strings.Index(h.Names[0], "#")
	if idx >= 0 {
		if idx >= len(*s)-1 {
			return nil, fmt.Errorf("No index after '#'")
		}
		i, err := strconv.Atoi(h.Names[0][(idx + 1):])
		if err != nil {
			return nil, err
		}
		h.Index = i
		substr := h.Names[0][0:(idx)]
		h.Names[0] = substr
	}
	return &h, nil
}

func main() {
	index := -1
	var flagTags string
	var flagRegion string
	var flagProfile string
	flag.Usage = usage
	flag.StringVar(&flagTags, "t", "", "Tags to filter the instances. TAG=VALUE,TAG=VALUE...")
	flag.StringVar(&flagRegion, "r", "", "AWS region to search in")
	flag.StringVar(&flagProfile, "p", "default", "Profile to use")
	flag.Parse()
	if flag.NArg() != 1 {
		usage()
		os.Exit(1)
	}
	hostName := flag.Args()[0]
	spec, err := parseHostName(&hostName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}
	index = spec.Index
	filters := make([]*ec2.Filter, 0)
	filters, err = parseTags(filters, &flagTags, spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}
	filters = addFilter(filters, "instance-state-name", "running")
	sessionOptions := session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Profile:           flagProfile,
	}
	if flagRegion != "" {
		region, err := replacePositional(flagRegion, spec)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Invalid region %s: %s\n", region, err.Error())
			os.Exit(1)
		}
		sessionOptions.Config = aws.Config{Region: aws.String(region)}
	}
	sess := session.Must(session.NewSessionWithOptions(sessionOptions))
	ec2Svc := ec2.New(sess)
	input := &ec2.DescribeInstancesInput{
		Filters: filters,
	}
	result, err := ec2Svc.DescribeInstances(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
	if len(result.Reservations) == 0 || len(result.Reservations[0].Instances) == 0 {
		fmt.Fprintf(os.Stderr, "Error: No matching host\n")
		os.Exit(1)
	}
	instances := make(EC2Instances, len(result.Reservations[0].Instances))
	copy(instances, result.Reservations[0].Instances)
	sort.Sort(instances)
	if index == -1 {
		if len(instances) > 1 {
			fmt.Fprintf(os.Stderr, "Error: %d hosts matches\n", len(instances))
			os.Exit(1)
		}
		index = 0
	} else {
		if len(instances) < index+1 {
			fmt.Fprintf(os.Stderr, "Error: There are only %d matching hosts\n", len(instances))
			os.Exit(1)
		}
	}
	fmt.Printf("%s\n", *instances[index].PublicDnsName)
}
