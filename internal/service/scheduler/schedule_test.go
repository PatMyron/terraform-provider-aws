package scheduler_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/scheduler"
	"github.com/aws/aws-sdk-go-v2/service/scheduler/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfscheduler "github.com/hashicorp/terraform-provider-aws/internal/service/scheduler"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestResourceScheduleIDFromARN(t *testing.T) {
	testCases := []struct {
		ARN   string
		ID    string
		Fails bool
	}{
		{
			ARN:   "arn:aws:scheduler:eu-west-1:735669964269:schedule/default/test",
			ID:    "default/test",
			Fails: false,
		},
		{
			ARN:   "arn:aws:scheduler:eu-west-1:735669964269:schedule/default/test/test",
			ID:    "",
			Fails: true,
		},
		{
			ARN:   "arn:aws:scheduler:eu-west-1:735669964269:schedule/default/",
			ID:    "",
			Fails: true,
		},
		{
			ARN:   "arn:aws:scheduler:eu-west-1:735669964269:schedule//test",
			ID:    "",
			Fails: true,
		},
		{
			ARN:   "arn:aws:scheduler:eu-west-1:735669964269:schedule//",
			ID:    "",
			Fails: true,
		},
		{
			ARN:   "arn:aws:scheduler:eu-west-1:735669964269:schedule/default",
			ID:    "",
			Fails: true,
		},
		{
			ARN:   "",
			ID:    "",
			Fails: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.ARN, func(t *testing.T) {
			id, err := tfscheduler.ResourceScheduleIDFromARN(tc.ARN)

			if tc.Fails {
				if err == nil {
					t.Errorf("expected an error")
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got: %s", err)
				}
			}

			if id != tc.ID {
				t.Errorf("expected id %s, got: %s", tc.ID, id)
			}
		})
	}
}

func TestResourceScheduleParseID(t *testing.T) {
	testCases := []struct {
		ID           string
		GroupName    string
		ScheduleName string
		Fails        bool
	}{
		{
			ID:           "default/test",
			GroupName:    "default",
			ScheduleName: "test",
			Fails:        false,
		},
		{
			ID:           "default/test/test",
			GroupName:    "",
			ScheduleName: "",
			Fails:        true,
		},
		{
			ID:           "default/",
			GroupName:    "",
			ScheduleName: "",
			Fails:        true,
		},
		{
			ID:           "/test",
			GroupName:    "",
			ScheduleName: "",
			Fails:        true,
		},
		{
			ID:           "/",
			GroupName:    "",
			ScheduleName: "",
			Fails:        true,
		},
		{
			ID:           "//",
			GroupName:    "",
			ScheduleName: "",
			Fails:        true,
		},
		{
			ID:           "default",
			GroupName:    "",
			ScheduleName: "",
			Fails:        true,
		},
		{
			ID:           "",
			GroupName:    "",
			ScheduleName: "",
			Fails:        true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.ID, func(t *testing.T) {
			groupName, scheduleName, err := tfscheduler.ResourceScheduleParseID(tc.ID)

			if tc.Fails {
				if err == nil {
					t.Errorf("expected an error")
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got: %s", err)
				}
			}

			if groupName != tc.GroupName {
				t.Errorf("expected group name %s, got: %s", tc.GroupName, groupName)
			}

			if scheduleName != tc.ScheduleName {
				t.Errorf("expected schedule name %s, got: %s", tc.ScheduleName, scheduleName)
			}
		})
	}
}

func TestAccSchedulerSchedule_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_scheduler_schedule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.SchedulerEndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "scheduler", regexp.MustCompile(regexp.QuoteMeta(`schedule/default/`+name))),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "end_date", ""),
					resource.TestCheckResourceAttr(resourceName, "flexible_time_window.0.maximum_window_in_minutes", "0"),
					resource.TestCheckResourceAttr(resourceName, "flexible_time_window.0.mode", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "group_name", "default"),
					resource.TestCheckResourceAttr(resourceName, "id", fmt.Sprintf("default/%s", name)),
					resource.TestCheckResourceAttr(resourceName, "kms_key_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "schedule_expression", "rate(1 hour)"),
					resource.TestCheckResourceAttr(resourceName, "schedule_expression_timezone", "UTC"),
					resource.TestCheckResourceAttr(resourceName, "start_date", ""),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttrPair(resourceName, "target.0.arn", "aws_sqs_queue.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "target.0.dead_letter_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target.0.eventbridge_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target.0.input", ""),
					resource.TestCheckResourceAttr(resourceName, "target.0.kinesis_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target.0.retry_policy.0.maximum_event_age_in_seconds", "86400"),
					resource.TestCheckResourceAttr(resourceName, "target.0.retry_policy.0.maximum_retry_attempts", "185"),
					resource.TestCheckResourceAttrPair(resourceName, "target.0.role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "target.0.sagemaker_pipeline_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "target.0.sqs_parameters.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSchedulerSchedule_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_scheduler_schedule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.SchedulerEndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					acctest.CheckResourceDisappears(acctest.Provider, tfscheduler.ResourceSchedule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSchedulerSchedule_description(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_scheduler_schedule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.SchedulerEndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_description(name, "test 1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "description", "test 1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_description(name, "test 2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "description", "test 2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_description(name, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSchedulerSchedule_endDate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_scheduler_schedule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.SchedulerEndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_endDate(name, "2100-01-01T01:02:03Z"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "end_date", "2100-01-01T01:02:03Z"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_endDate(name, "2099-01-01T01:00:00Z"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "end_date", "2099-01-01T01:00:00Z"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "end_date", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSchedulerSchedule_flexibleTimeWindow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_scheduler_schedule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.SchedulerEndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_flexibleTimeWindow(name, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "flexible_time_window.0.maximum_window_in_minutes", "10"),
					resource.TestCheckResourceAttr(resourceName, "flexible_time_window.0.mode", "FLEXIBLE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_flexibleTimeWindow(name, 20),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "flexible_time_window.0.maximum_window_in_minutes", "20"),
					resource.TestCheckResourceAttr(resourceName, "flexible_time_window.0.mode", "FLEXIBLE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "flexible_time_window.0.maximum_window_in_minutes", "0"),
					resource.TestCheckResourceAttr(resourceName, "flexible_time_window.0.mode", "OFF"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSchedulerSchedule_groupName(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_scheduler_schedule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.SchedulerEndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_groupName(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttrPair(resourceName, "group_name", "aws_scheduler_schedule_group.test", "name"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSchedulerSchedule_kmsKeyArn(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_scheduler_schedule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.SchedulerEndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_kmsKeyArn(name, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_arn", "aws_kms_key.test.0", "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_kmsKeyArn(name, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_arn", "aws_kms_key.test.1", "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "kms_key_arn", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSchedulerSchedule_nameGenerated(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	resourceName := "aws_scheduler_schedule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.SchedulerEndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_nameGenerated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					acctest.CheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", resource.UniqueIdPrefix),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSchedulerSchedule_namePrefix(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	resourceName := "aws_scheduler_schedule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.SchedulerEndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_namePrefix("tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, "name", "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "tf-acc-test-prefix-"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSchedulerSchedule_scheduleExpression(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_scheduler_schedule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.SchedulerEndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_scheduleExpression(name, "rate(1 hour)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "schedule_expression", "rate(1 hour)"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_scheduleExpression(name, "rate(1 day)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "schedule_expression", "rate(1 day)"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSchedulerSchedule_scheduleExpressionTimezone(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_scheduler_schedule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.SchedulerEndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_scheduleExpressionTimezone(name, "Europe/Paris"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "schedule_expression_timezone", "Europe/Paris"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_scheduleExpressionTimezone(name, "Australia/Sydney"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "schedule_expression_timezone", "Australia/Sydney"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "schedule_expression_timezone", "UTC"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSchedulerSchedule_startDate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_scheduler_schedule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.SchedulerEndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_startDate(name, "2100-01-01T01:02:03Z"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "start_date", "2100-01-01T01:02:03Z"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_startDate(name, "2099-01-01T01:00:00Z"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "start_date", "2099-01-01T01:00:00Z"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "start_date", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSchedulerSchedule_state(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_scheduler_schedule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.SchedulerEndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_state(name, "ENABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_state(name, "DISABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "state", "DISABLED"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSchedulerSchedule_targetArn(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_scheduler_schedule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.SchedulerEndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_targetArn(name, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttrPair(resourceName, "target.0.arn", "aws_sqs_queue.test.0", "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_targetArn(name, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttrPair(resourceName, "target.0.arn", "aws_sqs_queue.test.1", "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSchedulerSchedule_targetDeadLetterConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_scheduler_schedule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.SchedulerEndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_targetDeadLetterConfig(name, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttrPair(resourceName, "target.0.dead_letter_config.0.arn", "aws_sqs_queue.dlq.0", "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_targetDeadLetterConfig(name, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttrPair(resourceName, "target.0.dead_letter_config.0.arn", "aws_sqs_queue.dlq.1", "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "target.0.dead_letter_config.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSchedulerSchedule_targetEventBridgeParameters(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	scheduleName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	eventBusName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_scheduler_schedule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.SchedulerEndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_targetEventBridgeParameters(scheduleName, eventBusName, "test-1", "tf.test.1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "target.0.eventbridge_parameters.0.detail_type", "test-1"),
					resource.TestCheckResourceAttr(resourceName, "target.0.eventbridge_parameters.0.source", "tf.test.1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_targetEventBridgeParameters(scheduleName, eventBusName, "test-2", "tf.test.2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "target.0.eventbridge_parameters.0.detail_type", "test-2"),
					resource.TestCheckResourceAttr(resourceName, "target.0.eventbridge_parameters.0.source", "tf.test.2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_basic(scheduleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "target.0.eventbridge_parameters.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSchedulerSchedule_targetInput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_scheduler_schedule.test"
	var queueUrl string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.SchedulerEndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_targetInput(name, "test1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttrWith("aws_sqs_queue.test", "url", func(value string) error {
						queueUrl = value
						return nil
					}),
					func(s *terraform.State) error {
						return acctest.CheckResourceAttrEquivalentJSON(
							resourceName,
							"target.0.input",
							fmt.Sprintf(`{"MessageBody": "test1", "QueueUrl": %q}`, queueUrl),
						)(s)
					},
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_targetInput(name, "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					func(s *terraform.State) error {
						return acctest.CheckResourceAttrEquivalentJSON(
							resourceName,
							"target.0.input",
							fmt.Sprintf(`{"MessageBody": "test2", "QueueUrl": %q}`, queueUrl),
						)(s)
					},
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSchedulerSchedule_targetKinesisParameters(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	scheduleName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	streamName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_scheduler_schedule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.SchedulerEndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_targetKinesisParameters(scheduleName, streamName, "test-1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "target.0.kinesis_parameters.0.partition_key", "test-1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_targetKinesisParameters(scheduleName, streamName, "test-2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "target.0.kinesis_parameters.0.partition_key", "test-2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_basic(scheduleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "target.0.kinesis_parameters.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSchedulerSchedule_targetRetryPolicy(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_scheduler_schedule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.SchedulerEndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_targetRetryPolicy(name, 60, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "target.0.retry_policy.0.maximum_event_age_in_seconds", "60"),
					resource.TestCheckResourceAttr(resourceName, "target.0.retry_policy.0.maximum_retry_attempts", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_targetRetryPolicy(name, 61, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "target.0.retry_policy.0.maximum_event_age_in_seconds", "61"),
					resource.TestCheckResourceAttr(resourceName, "target.0.retry_policy.0.maximum_retry_attempts", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "target.0.retry_policy.0.maximum_event_age_in_seconds", "86400"),
					resource.TestCheckResourceAttr(resourceName, "target.0.retry_policy.0.maximum_retry_attempts", "185"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSchedulerSchedule_targetRoleArn(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_scheduler_schedule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.SchedulerEndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_targetRoleArn(name, "test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttrPair(resourceName, "target.0.role_arn", "aws_iam_role.test", "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_targetRoleArn(name, "test1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttrPair(resourceName, "target.0.role_arn", "aws_iam_role.test1", "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSchedulerSchedule_targetSageMakerPipelineParameters(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_scheduler_schedule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.SchedulerEndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_targetSageMakerPipelineParameters1(name, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "target.0.sagemaker_pipeline_parameters.0.pipeline_parameter.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(
						resourceName, "target.0.sagemaker_pipeline_parameters.0.pipeline_parameter.*",
						map[string]string{
							"name":  "key1",
							"value": "value1",
						}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_targetSageMakerPipelineParameters2(name, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "target.0.sagemaker_pipeline_parameters.0.pipeline_parameter.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(
						resourceName, "target.0.sagemaker_pipeline_parameters.0.pipeline_parameter.*",
						map[string]string{
							"name":  "key1",
							"value": "value1updated",
						}),
					resource.TestCheckTypeSetElemNestedAttrs(
						resourceName, "target.0.sagemaker_pipeline_parameters.0.pipeline_parameter.*",
						map[string]string{
							"name":  "key2",
							"value": "value2",
						}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_targetSageMakerPipelineParameters1(name, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "target.0.sagemaker_pipeline_parameters.0.pipeline_parameter.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(
						resourceName, "target.0.sagemaker_pipeline_parameters.0.pipeline_parameter.*",
						map[string]string{
							"name":  "key2",
							"value": "value2",
						}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "target.0.sagemaker_pipeline_parameters.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSchedulerSchedule_targetSqsParameters(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_scheduler_schedule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.SchedulerEndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_targetSqsParameters(name, "test1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "target.0.sqs_parameters.0.message_group_id", "test1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_targetSqsParameters(name, "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "target.0.sqs_parameters.0.message_group_id", "test2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "target.0.sqs_parameters.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckScheduleDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SchedulerClient
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_scheduler_schedule" {
			continue
		}

		parts := strings.Split(rs.Primary.ID, "/")

		input := &scheduler.GetScheduleInput{
			GroupName: aws.String(parts[0]),
			Name:      aws.String(parts[1]),
		}
		_, err := conn.GetSchedule(ctx, input)
		if err != nil {
			var nfe *types.ResourceNotFoundException
			if errors.As(err, &nfe) {
				return nil
			}
			return err
		}

		return create.Error(names.Scheduler, create.ErrActionCheckingDestroyed, tfscheduler.ResNameSchedule, rs.Primary.ID, errors.New("not destroyed"))
	}

	return nil
}

func testAccCheckScheduleExists(name string, schedule *scheduler.GetScheduleOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Scheduler, create.ErrActionCheckingExistence, tfscheduler.ResNameSchedule, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Scheduler, create.ErrActionCheckingExistence, tfscheduler.ResNameSchedule, name, errors.New("not set"))
		}

		parts := strings.Split(rs.Primary.ID, "/")

		conn := acctest.Provider.Meta().(*conns.AWSClient).SchedulerClient
		ctx := context.Background()
		resp, err := conn.GetSchedule(ctx, &scheduler.GetScheduleInput{
			Name:      aws.String(parts[1]),
			GroupName: aws.String(parts[0]),
		})

		if err != nil {
			return create.Error(names.Scheduler, create.ErrActionCheckingExistence, tfscheduler.ResNameSchedule, rs.Primary.ID, err)
		}

		*schedule = *resp

		return nil
	}
}

const testAccScheduleConfig_base = `
data "aws_caller_identity" "main" {}
data "aws_partition" "main" {}

resource "aws_iam_role" "test" {
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = {
      Effect = "Allow"
      Action = "sts:AssumeRole"
      Principal = {
        Service = "scheduler.${data.aws_partition.main.dns_suffix}"
      }
      Condition = {
        StringEquals = {
          "aws:SourceAccount": data.aws_caller_identity.main.account_id
        }
      }
    }
  })
}
`

func testAccScheduleConfig_basic(name string) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
resource "aws_sqs_queue" "test" {}

resource "aws_scheduler_schedule" "test" {
  name = %[1]q

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = "rate(1 hour)"

  target {
    arn      = aws_sqs_queue.test.arn
    role_arn = aws_iam_role.test.arn
  }
}
`, name),
	)
}

func testAccScheduleConfig_description(name, description string) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
resource "aws_sqs_queue" "test" {}

resource "aws_scheduler_schedule" "test" {
  name = %[1]q

  description = %[2]q

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = "rate(1 hour)"

  target {
    arn      = aws_sqs_queue.test.arn
    role_arn = aws_iam_role.test.arn
  }
}
`, name, description),
	)
}

func testAccScheduleConfig_endDate(name, endDate string) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
resource "aws_sqs_queue" "test" {}

resource "aws_scheduler_schedule" "test" {
  name = %[1]q

  end_date = %[2]q

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = "rate(1 hour)"

  target {
    arn      = aws_sqs_queue.test.arn
    role_arn = aws_iam_role.test.arn
  }
}
`, name, endDate),
	)
}

func testAccScheduleConfig_flexibleTimeWindow(name string, window int) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
resource "aws_sqs_queue" "test" {}

resource "aws_scheduler_schedule" "test" {
  name = %[1]q

  flexible_time_window {
    maximum_window_in_minutes = %[2]d
    mode                      = "FLEXIBLE"
  }

  schedule_expression = "rate(1 hour)"

  target {
    arn      = aws_sqs_queue.test.arn
    role_arn = aws_iam_role.test.arn
  }
}
`, name, window),
	)
}

func testAccScheduleConfig_groupName(name string) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
resource "aws_sqs_queue" "test" {}

resource "aws_scheduler_schedule_group" "test" {}

resource "aws_scheduler_schedule" "test" {
  name = %[1]q

  flexible_time_window {
    mode = "OFF"
  }

  group_name = aws_scheduler_schedule_group.test.name

  schedule_expression = "rate(1 hour)"

  target {
    arn      = aws_sqs_queue.test.arn
    role_arn = aws_iam_role.test.arn
  }
}
`, name),
	)
}

func testAccScheduleConfig_kmsKeyArn(name string, index int) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
resource "aws_sqs_queue" "test" {}

resource "aws_kms_key" "test" {
  count = 2
}

resource "aws_scheduler_schedule" "test" {
  name = %[1]q

  flexible_time_window {
    mode = "OFF"
  }

  kms_key_arn = aws_kms_key.test[%[2]d].arn

  schedule_expression = "rate(1 hour)"

  target {
    arn      = aws_sqs_queue.test.arn
    role_arn = aws_iam_role.test.arn
  }
}
`, name, index),
	)
}

func testAccScheduleConfig_nameGenerated() string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		`
resource "aws_sqs_queue" "test" {}

resource "aws_scheduler_schedule" "test" {
  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = "rate(1 hour)"

  target {
    arn      = aws_sqs_queue.test.arn
    role_arn = aws_iam_role.test.arn
  }
}
`,
	)
}

func testAccScheduleConfig_namePrefix(namePrefix string) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
resource "aws_sqs_queue" "test" {}

resource "aws_scheduler_schedule" "test" {
  name_prefix = %[1]q

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = "rate(1 hour)"

  target {
    arn      = aws_sqs_queue.test.arn
    role_arn = aws_iam_role.test.arn
  }
}
`, namePrefix),
	)
}

func testAccScheduleConfig_scheduleExpression(name, expression string) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
resource "aws_sqs_queue" "test" {}

resource "aws_scheduler_schedule" "test" {
  name = %[1]q

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = %[2]q

  target {
    arn      = aws_sqs_queue.test.arn
    role_arn = aws_iam_role.test.arn
  }
}
`, name, expression),
	)
}

func testAccScheduleConfig_scheduleExpressionTimezone(name, timezone string) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
resource "aws_sqs_queue" "test" {}

resource "aws_scheduler_schedule" "test" {
  name = %[1]q

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = "rate(1 hour)"

  schedule_expression_timezone = %[2]q

  target {
    arn      = aws_sqs_queue.test.arn
    role_arn = aws_iam_role.test.arn
  }
}
`, name, timezone),
	)
}

func testAccScheduleConfig_startDate(name, startDate string) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
resource "aws_sqs_queue" "test" {}

resource "aws_scheduler_schedule" "test" {
  name = %[1]q

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = "rate(1 hour)"

  start_date = %[2]q

  target {
    arn      = aws_sqs_queue.test.arn
    role_arn = aws_iam_role.test.arn
  }
}
`, name, startDate),
	)
}

func testAccScheduleConfig_state(name, state string) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
resource "aws_sqs_queue" "test" {}

resource "aws_scheduler_schedule" "test" {
  name = %[1]q

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = "rate(1 hour)"

  state = %[2]q

  target {
    arn      = aws_sqs_queue.test.arn
    role_arn = aws_iam_role.test.arn
  }
}
`, name, state),
	)
}

func testAccScheduleConfig_targetArn(name string, i int) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  count = 2
}

resource "aws_scheduler_schedule" "test" {
  name = %[1]q

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = "rate(1 hour)"

  target {
    arn      = aws_sqs_queue.test[%[2]d].arn
    role_arn = aws_iam_role.test.arn
  }
}
`, name, i),
	)
}

func testAccScheduleConfig_targetDeadLetterConfig(name string, index int) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
resource "aws_sqs_queue" "test" {}

resource "aws_sqs_queue" "dlq" {
	count = 2
}

resource "aws_scheduler_schedule" "test" {
  name = %[1]q

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = "rate(1 hour)"

  target {
    arn      = aws_sqs_queue.test.arn
    role_arn = aws_iam_role.test.arn

    dead_letter_config {
      arn = aws_sqs_queue.dlq[%[2]d].arn
    }
  }
}
`, name, index),
	)
}

func testAccScheduleConfig_targetEventBridgeParameters(scheduleName, eventBusName, detailType, source string) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name = %[2]q
}

resource "aws_scheduler_schedule" "test" {
  name = %[1]q

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = "rate(1 hour)"

  target {
    arn      = aws_cloudwatch_event_bus.test.arn
    role_arn = aws_iam_role.test.arn

    eventbridge_parameters {
      detail_type = %[3]q
      source      = %[4]q
    }
  }
}
`, scheduleName, eventBusName, detailType, source),
	)
}

func testAccScheduleConfig_targetInput(name, messageBody string) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
resource "aws_sqs_queue" "test" {}

resource "aws_scheduler_schedule" "test" {
  name = %[1]q

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = "rate(1 hour)"

  target {
    arn      = "arn:aws:scheduler:::aws-sdk:sqs:sendMessage"
    role_arn = aws_iam_role.test.arn

    input = jsonencode({
      MessageBody = %[2]q
      QueueUrl    = aws_sqs_queue.test.url
    })
  }
}
`, name, messageBody),
	)
}

func testAccScheduleConfig_targetKinesisParameters(scheduleName, streamName, partitionKey string) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  name        = %[2]q
  shard_count = 1
}

resource "aws_scheduler_schedule" "test" {
  name = %[1]q

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = "rate(1 hour)"

  target {
    arn      = aws_kinesis_stream.test.arn
    role_arn = aws_iam_role.test.arn

    kinesis_parameters {
      partition_key = %[3]q
    }
  }
}
`, scheduleName, streamName, partitionKey),
	)
}

func testAccScheduleConfig_targetRetryPolicy(name string, maxEventAge, maxRetryAttempts int) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
resource "aws_sqs_queue" "test" {}

resource "aws_sqs_queue" "dlq" {
	count = 2
}

resource "aws_scheduler_schedule" "test" {
  name = %[1]q

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = "rate(1 hour)"

  target {
    arn      = aws_sqs_queue.test.arn
    role_arn = aws_iam_role.test.arn

    retry_policy {
      maximum_event_age_in_seconds = %[2]d
      maximum_retry_attempts       = %[3]d
    }
  }
}
`, name, maxEventAge, maxRetryAttempts),
	)
}

func testAccScheduleConfig_targetRoleArn(name, resourceName string) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
resource "aws_sqs_queue" "test" {}

resource "aws_iam_role" "test1" {
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = {
      Effect = "Allow"
      Action = "sts:AssumeRole"
      Principal = {
        Service = "scheduler.${data.aws_partition.main.dns_suffix}"
      }
    }
  })
}

resource "aws_scheduler_schedule" "test" {
  name = %[1]q

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = "rate(1 hour)"

  target {
    arn      = aws_sqs_queue.test.arn
    role_arn = aws_iam_role.%[2]s.arn
  }
}
`, name, resourceName),
	)
}

func testAccScheduleConfig_targetSageMakerPipelineParameters1(name, name1, value1 string) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
data "aws_region" "main" {}

resource "aws_scheduler_schedule" "test" {
  name = %[1]q

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = "rate(1 hour)"

  target {
    arn      = "arn:aws:sagemaker:${data.aws_region.main.name}:${data.aws_caller_identity.main.account_id}:pipeline/test"
    role_arn = aws_iam_role.test.arn

    sagemaker_pipeline_parameters {
      pipeline_parameter {
        name  = %[2]q
        value = %[3]q
      }
    }
  }
}
`, name, name1, value1),
	)
}

func testAccScheduleConfig_targetSageMakerPipelineParameters2(name, name1, value1, name2, value2 string) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
data "aws_region" "main" {}

resource "aws_scheduler_schedule" "test" {
  name = %[1]q

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = "rate(1 hour)"

  target {
    arn      = "arn:aws:sagemaker:${data.aws_region.main.name}:${data.aws_caller_identity.main.account_id}:pipeline/test"
    role_arn = aws_iam_role.test.arn

    sagemaker_pipeline_parameters {
      pipeline_parameter {
        name  = %[2]q
        value = %[3]q
      }

      pipeline_parameter {
        name  = %[4]q
        value = %[5]q
      }
    }
  }
}
`, name, name1, value1, name2, value2),
	)
}

func testAccScheduleConfig_targetSqsParameters(name, messageGroupId string) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
resource "aws_sqs_queue" "test" {}

resource "aws_sqs_queue" "fifo" {
  fifo_queue = true
}

resource "aws_scheduler_schedule" "test" {
  name = %[1]q

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = "rate(1 hour)"

  target {
    arn      = aws_sqs_queue.fifo.arn
    role_arn = aws_iam_role.test.arn

    sqs_parameters {
      message_group_id = %[2]q
    }
  }
}
`, name, messageGroupId),
	)
}
