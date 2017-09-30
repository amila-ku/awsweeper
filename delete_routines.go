package main

import (
	"reflect"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/wafregional"
	"github.com/aws/aws-sdk-go/service/cloudtrail"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/aws/aws-sdk-go/service/cognitoidentity"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/aws/aws-sdk-go/service/devicefarm"
	"github.com/aws/aws-sdk-go/service/databasemigrationservice"
	"github.com/aws/aws-sdk-go/service/directoryservice"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/iam"
)

type AWSClient struct {
	cfconn                *cloudformation.CloudFormation
	cloudfrontconn        *cloudfront.CloudFront
	cloudtrailconn        *cloudtrail.CloudTrail
	cloudwatchconn        *cloudwatch.CloudWatch
	cloudwatchlogsconn    *cloudwatchlogs.CloudWatchLogs
	cloudwatcheventsconn  *cloudwatchevents.CloudWatchEvents
	cognitoconn           *cognitoidentity.CognitoIdentity
	configconn            *configservice.ConfigService
	devicefarmconn        *devicefarm.DeviceFarm
	dmsconn               *databasemigrationservice.DatabaseMigrationService
	dsconn                *directoryservice.DirectoryService
	dynamodbconn          *dynamodb.DynamoDB
	ec2conn               *ec2.EC2
	ecrconn               *ecr.ECR
	ecsconn               *ecs.ECS
	efsconn               *efs.EFS
	elbconn               *elb.ELB
	elbv2conn             *elbv2.ELBV2
	emrconn               *emr.EMR
	esconn                *elasticsearch.ElasticsearchService
	acmconn               *acm.ACM
	apigateway            *apigateway.APIGateway
	appautoscalingconn    *applicationautoscaling.ApplicationAutoScaling
	autoscalingconn       *autoscaling.AutoScaling
	s3conn                *s3.S3
	sesConn               *ses.SES
	simpledbconn          *simpledb.SimpleDB
	sqsconn               *sqs.SQS
	snsconn               *sns.SNS
	stsconn               *sts.STS
	redshiftconn          *redshift.Redshift
	r53conn               *route53.Route53
	partition             string
	accountid             string
	supportedplatforms    []string
	region                string
	rdsconn               *rds.RDS
	iamconn               *iam.IAM
	kinesisconn           *kinesis.Kinesis
	kmsconn               *kms.KMS
	firehoseconn          *firehose.Firehose
	inspectorconn         *inspector.Inspector
	elasticacheconn       *elasticache.ElastiCache
	elasticbeanstalkconn  *elasticbeanstalk.ElasticBeanstalk
	elastictranscoderconn *elastictranscoder.ElasticTranscoder
	lambdaconn            *lambda.Lambda
	lightsailconn         *lightsail.Lightsail
	opsworksconn          *opsworks.OpsWorks
	glacierconn           *glacier.Glacier
	codebuildconn         *codebuild.CodeBuild
	codedeployconn        *codedeploy.CodeDeploy
	codecommitconn        *codecommit.CodeCommit
	codepipelineconn      *codepipeline.CodePipeline
	sfnconn               *sfn.SFN
	ssmconn               *ssm.SSM
	wafconn               *waf.WAF
	wafregionalconn       *wafregional.WAFRegional
}

type ResourceSet struct {
	Type  string
	Ids   []*string
	Attrs []*map[string]string
	Tags  []*map[string]string
	Info  []string
}

type DeleteResourceInfo struct {
	TerraformType string
	DescribeOutputName string
	DeleteId string
	DescribeFn interface{}
	DescribeFnInput interface{}
}

func listResources(fn interface{}, args ...interface{}) interface {} {
	v := reflect.ValueOf(fn)
	rargs := make([]reflect.Value, len(args))
	for i, a := range args {
		rargs[i] = reflect.ValueOf(a)
	}
	result := v.Call(rargs)
	return result[0].Interface()
}

type DeleteRoutines struct {
	deleteInfo      []DeleteResourceInfo
}

func Bla(c AWSClient) DeleteRoutines {
	return DeleteRoutines{
		deleteInfo: []DeleteResourceInfo{
			{"aws_autoscaling_group", "AutoScalingGroups", "AutoScalingGroupName",
			 c.autoscalingconn.DescribeAutoScalingGroups, &autoscaling.DescribeAutoScalingGroupsInput{}},
			{"aws_launch_configuration", "LaunchConfigurations", "LaunchConfigurationName",
			 c.autoscalingconn.DescribeLaunchConfigurations, &autoscaling.DescribeLaunchConfigurationsInput{}},
			//"aws_instance":             c.deleteInstances, // c.deleteInstances
			{"aws_internet_gateway", "InternetGateways", "InternetGatewayId",
			 c.ec2conn.DescribeInternetGateways, &ec2.DescribeInternetGatewaysInput{}},
			{"aws_eip", "Addresses", "AllocationId",
			 c.ec2conn.DescribeAddresses, &ec2.DescribeAddressesInput{}},
			{"aws_elb", "LoadBalancerDescriptions", "LoadBalancerName",
			 c.elbconn.DescribeLoadBalancers, &elb.DescribeLoadBalancersInput{}},
			{"aws_vpc_endpoint", "VpcEndpoints", "VpcEndpointId",
			 c.ec2conn.DescribeVpcEndpoints, &ec2.DescribeVpcEndpointsInput{}},
			{"aws_nat_gateway", "NatGateways", "NatGatewayId",
			 &ec2.DescribeNatGatewaysInput{}, &ec2.DescribeNatGatewaysInput{}},
			{"aws_network_interface", "NetworkInterfaces", "NetworkInterfaceId",
			 c.ec2conn.DescribeNetworkInterfaces, &ec2.DescribeNetworkInterfacesInput{}},
			{"aws_route_table", "RouteTables", "RouteTableId",
			 &ec2.DescribeRouteTablesInput{}, c.ec2conn.DescribeRouteTables},
			{"aws_security_group", "SecurityGroups", "GroupId",
			 c.ec2conn.DescribeSecurityGroups, &ec2.DescribeSecurityGroupsInput{}},
			{"aws_network_acl", "NetworkAcls", "NetworkAclId",
			 &ec2.DescribeNetworkAclsInput{}, &ec2.DescribeNetworkAclsInput{}},
			{"aws_subnet", "Subnets", "SubnetId",
			 c.ec2conn.DescribeSubnets, &ec2.DescribeSubnetsInput{}},
			// TODO filter by name?
			{"aws_cloudformation_stack", "Stacks", "StackId",
			 c.cfconn.DescribeStacks, &cloudformation.DescribeStacksInput{}},
			//{"aws_route53_zone", "HostedZones", "Id"},
			{"aws_vpc", "Vpcs", "VpcId",
			 c.ec2conn.DescribeVpcs, &ec2.DescribeVpcsInput{}},
			//{"aws_efs_file_system", "FileSystems", "FileSystemId"},
			//{"aws_iam_user", "Users", "UserName"},
			//{"aws_iam_role", "Roles", "RoleName"},
			//{"aws_iam_policy", "Policies", "Arn"},
			{"aws_iam_instance_profile", "InstanceProfiles", "InstanceProfileName",
			 c.iamconn.ListInstanceProfiles, &iam.ListInstanceProfilesInput{}},
			//{"aws_kms_alias", "Aliases", "AliasName"}, //  c.deleteKmsAliases
			//{"aws_kms_key", "Keys", "KeyId"}, // c.deleteKmsKeys
			{"aws_ebs_snapshot", "Snapshots", "SnapshotId",
			 c.ec2conn.DescribeSnapshots, &ec2.DescribeSnapshotsInput{}},
			// TODO filter by name?
			{"aws_ebs_volume", "Volumes", "VolumeId",
			 c.ec2conn.DescribeVolumes, &ec2.DescribeVolumesInput{}},
			//{"aws_ami", "Image
		},
	}
}
