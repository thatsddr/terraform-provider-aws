package mediatailor_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/mediatailor"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccChannelResource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_media_tailor_channel.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, mediatailor.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckChannelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccChannelConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "channel_name", rName),
					resource.TestCheckResourceAttr(resourceName, "outputs.0.manifest_name", "default"),
					resource.TestCheckResourceAttr(resourceName, "outputs.0.source_group", "default"),
					resource.TestCheckResourceAttr(resourceName, "outputs.0.hls_manifest_windows_seconds", "30"),
					resource.TestMatchResourceAttr(resourceName, "outputs.0.playback_url", regexp.MustCompile(`^https:\/\/[\w+.\/-]+.(mpd|m3u8)$`)),
					resource.TestCheckResourceAttr(resourceName, "playback_mode", "LOOP"),
					resource.TestCheckResourceAttr(resourceName, "tier", "BASIC"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateVerify: true,
				ImportState:       true,
			},
		},
	})
}

func TestAccChannelResource_recreate(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_media_tailor_channel.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, mediatailor.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckChannelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccChannelConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "channel_name", rName),
					resource.TestCheckResourceAttr(resourceName, "outputs.0.manifest_name", "default"),
					resource.TestCheckResourceAttr(resourceName, "outputs.0.source_group", "default"),
					resource.TestCheckResourceAttr(resourceName, "outputs.0.hls_manifest_windows_seconds", "30"),
					resource.TestMatchResourceAttr(resourceName, "outputs.0.playback_url", regexp.MustCompile(`^https:\/\/[\w+.\/-]+.(mpd|m3u8)$`)),
					resource.TestCheckResourceAttr(resourceName, "playback_mode", "LOOP"),
					resource.TestCheckResourceAttr(resourceName, "tier", "BASIC"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "0"),
				),
			},
			{
				Taint:  []string{resourceName},
				Config: testAccChannelConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "channel_name", rName),
					resource.TestCheckResourceAttr(resourceName, "outputs.0.manifest_name", "default"),
					resource.TestCheckResourceAttr(resourceName, "outputs.0.source_group", "default"),
					resource.TestCheckResourceAttr(resourceName, "outputs.0.hls_manifest_windows_seconds", "30"),
					resource.TestMatchResourceAttr(resourceName, "outputs.0.playback_url", regexp.MustCompile(`^https:\/\/[\w+.\/-]+.(mpd|m3u8)$`)),
					resource.TestCheckResourceAttr(resourceName, "playback_mode", "LOOP"),
					resource.TestCheckResourceAttr(resourceName, "tier", "BASIC"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "0"),
				),
			},
		},
	})
}

func TestAccChannelResource_conflict(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, mediatailor.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckChannelDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccChannelConfig_Conflict(rName),
				ExpectError: regexp.MustCompile(regexp.QuoteMeta("The channel isn't valid. Every output must have exactly one of the DashPlaylistSettings attribute or the HlsPlaylistSettings attribute")),
			},
		},
	})
}

func TestAccChannelResource_validateTier(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, mediatailor.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckChannelDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccChannelConfig_Tier(rName, "TEST"),
				ExpectError: regexp.MustCompile(regexp.QuoteMeta("expected tier to be one of [BASIC STANDARD]")),
			},
		},
	})
}

func TestAccChannelResource_validatePlaybackMode(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, mediatailor.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckChannelDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccChannelConfig_PlaybackMode(rName, "TEST"),
				ExpectError: regexp.MustCompile(regexp.QuoteMeta("expected playback_mode to be one of [LINEAR LOOP]")),
			},
		},
	})
}

func TestAccChannelResource_update(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_media_tailor_channel.test"
	number := 30
	updatedNumber := 35
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, mediatailor.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckChannelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccChannelConfig_Update(rName, number),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "channel_name", rName),
					resource.TestCheckResourceAttr(resourceName, "outputs.0.dash_manifest_windows_seconds", fmt.Sprint(number)),
					resource.TestCheckResourceAttr(resourceName, "outputs.0.dash_min_buffer_time_seconds", fmt.Sprint(number)),
					resource.TestCheckResourceAttr(resourceName, "outputs.0.dash_min_update_period_seconds", fmt.Sprint(number)),
					resource.TestCheckResourceAttr(resourceName, "outputs.0.dash_suggested_presentation_delay_seconds", fmt.Sprint(number)),
					resource.TestMatchResourceAttr(resourceName, "outputs.0.playback_url", regexp.MustCompile(`^https:\/\/[\w+.\/-]+.(mpd|m3u8)$`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "0"),
				),
			},
			{
				Config: testAccChannelConfig_Update(rName, updatedNumber),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "channel_name", rName),
					resource.TestCheckResourceAttr(resourceName, "outputs.0.dash_manifest_windows_seconds", fmt.Sprint(updatedNumber)),
					resource.TestCheckResourceAttr(resourceName, "outputs.0.dash_min_buffer_time_seconds", fmt.Sprint(updatedNumber)),
					resource.TestCheckResourceAttr(resourceName, "outputs.0.dash_min_update_period_seconds", fmt.Sprint(updatedNumber)),
					resource.TestCheckResourceAttr(resourceName, "outputs.0.dash_suggested_presentation_delay_seconds", fmt.Sprint(updatedNumber)),
					resource.TestMatchResourceAttr(resourceName, "outputs.0.playback_url", regexp.MustCompile(`^https:\/\/[\w+.\/-]+.(mpd|m3u8)$`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "0"),
				),
			},
		},
	})
}

func testAccCheckChannelDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).MediaTailorConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_media_tailor_channel" {
			continue
		}

		var resourceName string

		if arn.IsARN(rs.Primary.ID) {
			resourceArn, err := arn.Parse(rs.Primary.ID)
			if err != nil {
				return fmt.Errorf("error parsing resource arn: %s.\n%s", err, rs.Primary.ID)
			}
			arnSections := strings.Split(resourceArn.Resource, "/")
			resourceName = arnSections[len(arnSections)-1]
		} else {
			resourceName = rs.Primary.ID
		}

		input := &mediatailor.DescribeChannelInput{ChannelName: aws.String(resourceName)}
		_, err := conn.DescribeChannel(input)

		if tfawserr.ErrCodeContains(err, "NotFound") {
			continue
		}

		if err != nil {
			return err
		}
	}
	return nil
}

func testAccChannelConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_media_tailor_channel" "test"{
  channel_name = "%[1]s"
  outputs {
    manifest_name                = "default"
    source_group                 = "default"
    hls_manifest_windows_seconds = 30
  }
  playback_mode = "LOOP"
  tier = "BASIC"
}
`, rName)
}

func testAccChannelConfig_Conflict(rName string) string {
	return fmt.Sprintf(`
resource "aws_media_tailor_channel" "test"{
  channel_name = "%[1]s"
  outputs {
    manifest_name                 = "default"
    source_group                  = "default"
	dash_manifest_windows_seconds = 30
    hls_manifest_windows_seconds  = 30
  }
  playback_mode = "LOOP"
  tier = "BASIC"
}
`, rName)
}

func testAccChannelConfig_Tier(rName, tier string) string {
	return fmt.Sprintf(`
resource "aws_media_tailor_channel" "test"{
  channel_name = "%[1]s"
  outputs {
    manifest_name                 = "default"
    source_group                  = "default"
	dash_manifest_windows_seconds = 30
  }
  playback_mode = "LOOP"
  tier = "%[2]s"
}
`, rName, tier)
}

func testAccChannelConfig_PlaybackMode(rName, playbackMode string) string {
	return fmt.Sprintf(`
resource "aws_media_tailor_channel" "test"{
  channel_name = "%[1]s"
  outputs {
    manifest_name                 = "default"
    source_group                  = "default"
	dash_manifest_windows_seconds = 30
  }
  playback_mode = "%[2]s"
  tier = "BASIC"
}
`, rName, playbackMode)
}

func testAccChannelConfig_Update(rName string, num int) string {
	return fmt.Sprintf(`
resource "aws_media_tailor_channel" "test"{
  channel_name = "%[1]s"
  outputs {
    manifest_name                             = "default"
    source_group                              = "default"
	dash_manifest_windows_seconds             = %[2]d
    dash_min_buffer_time_seconds              = %[2]d
    dash_min_update_period_seconds            = %[2]d
    dash_suggested_presentation_delay_seconds = %[2]d
  }
  playback_mode = "LOOP"
  tier = "BASIC"
}
`, rName, num)
}
