package mediatailor

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/mediatailor"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourcePlaybackConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePlaybackConfigurationCreate,
		ReadContext:   resourcePlaybackConfigurationRead,
		UpdateContext: resourcePlaybackConfigurationUpdate,
		DeleteContext: resourcePlaybackConfigurationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"ad_decision_server_url": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 25000),
			},
			"avail_suppression_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"OFF", "BEHIND_LIVE_EDGE"}, false),
			},
			"avail_suppression_value": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^\d{2}:\d{2}:\d{2}$`), "must be valid HH:MM:SS string"),
			},
			"bumper": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"end_url": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"start_url": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"cdn_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ad_segment_url_prefix": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"content_segment_url_prefix": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"dash_mpd_location": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"DISABLED", "EMT_DEFAULT"}, false),
			},
			"dash_origin_manifest_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"SINGLE_PERIOD", "MULTI_PERIOD"}, false),
			},
			"dash_manifest_endpoint_prefix": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hls_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"manifest_endpoint_prefix": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"live_pre_roll_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ad_decision_server_url": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 25000),
						},
						"max_duration_seconds": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      3600,
							ValidateFunc: validation.IntAtLeast(1),
						},
					},
				},
			},
			"log_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"percent_enabled": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"manifest_processing_rules": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ad_marker_passthrough": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enabled": {
										Type:     schema.TypeBool,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"personalization_threshold_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"playback_configuration_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"playback_endpoint_prefix": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"session_initialization_endpoint_prefix": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"slate_ad_url": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"transcode_profile_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"video_content_source_url": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 512),
			},
		},
		CustomizeDiff: customdiff.Sequence(
			customdiff.ForceNewIfChange("name", func(ctx context.Context, old, new, meta interface{}) bool { return old.(string) != new.(string) }),
			verify.SetTagsDiff,
		),
	}
}

func resourcePlaybackConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MediaTailorConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	var params = getPlaybackConfigurationInput(d)

	if len(tags) > 0 {
		params.Tags = Tags(tags.IgnoreAWS())
	}

	playbackConfiguration, err := conn.PutPlaybackConfiguration(&params)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error while creating the playback configuration: %v", err))
	}

	d.SetId(aws.StringValue(playbackConfiguration.PlaybackConfigurationArn))

	return resourcePlaybackConfigurationRead(ctx, d, meta)
}

func resourcePlaybackConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MediaTailorConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(conn, d.Get("playback_configuration_arn").(string), o, n); err != nil {
			return diag.FromErr(fmt.Errorf("error updating tags: %w", err))
		}
	}

	var params = getPlaybackConfigurationInput(d)
	playbackConfiguration, err := conn.PutPlaybackConfiguration(&params)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error while creating the playback configuration: %v", err))
	}
	d.SetId(aws.StringValue(playbackConfiguration.PlaybackConfigurationArn))

	return resourcePlaybackConfigurationRead(ctx, d, meta)
}

func resourcePlaybackConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MediaTailorConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	resourceName := d.Get("name").(string)
	if len(resourceName) == 0 && len(d.Id()) > 0 {
		resourceArn, err := arn.Parse(d.Id())
		if err != nil {
			return diag.FromErr(fmt.Errorf("error parsing the name from resource arn: %v", err))
		}
		arnSections := strings.Split(resourceArn.Resource, "/")
		resourceName = arnSections[len(arnSections)-1]
	}

	res, err := conn.GetPlaybackConfiguration(&mediatailor.GetPlaybackConfigurationInput{Name: aws.String(resourceName)})
	if err != nil {
		return diag.FromErr(fmt.Errorf("error while retrieving the resource: %v", err))
	}

	d.Set("ad_decision_server_url", res.AdDecisionServerUrl)
	d.Set("avail_suppression_mode", res.AvailSuppression.Mode)
	d.Set("avail_suppression_value", res.AvailSuppression.Value)

	if res.Bumper.StartUrl != nil || res.Bumper.EndUrl != nil {
		temp := map[string]interface{}{}
		if res.Bumper.StartUrl != nil {
			temp["end_url"] = res.Bumper.EndUrl
		}
		if res.Bumper.StartUrl != nil {
			temp["start_url"] = res.Bumper.StartUrl
		}
		d.Set("bumper", []interface{}{temp})
	}

	if res.CdnConfiguration.AdSegmentUrlPrefix != nil || res.CdnConfiguration.ContentSegmentUrlPrefix != nil {
		temp := map[string]interface{}{}
		if res.CdnConfiguration.AdSegmentUrlPrefix != nil {
			temp["ad_segment_url_prefix"] = res.CdnConfiguration.AdSegmentUrlPrefix
		}
		if res.CdnConfiguration.ContentSegmentUrlPrefix != nil {
			temp["content_segment_url_prefix"] = res.CdnConfiguration.ContentSegmentUrlPrefix
		}
		d.Set("cdn_configuration", []interface{}{temp})
	}

	d.Set("dash_mpd_location", res.DashConfiguration.MpdLocation)
	d.Set("dash_origin_manifest_type", res.DashConfiguration.OriginManifestType)
	d.Set("dash_manifest_endpoint_prefix", res.DashConfiguration.ManifestEndpointPrefix)

	if res.HlsConfiguration != nil {
		temp := map[string]interface{}{}
		if res.HlsConfiguration.ManifestEndpointPrefix != nil {
			temp["manifest_endpoint_prefix"] = res.HlsConfiguration.ManifestEndpointPrefix
		}
		d.Set("hls_configuration", []interface{}{temp})
	}

	if res.LivePreRollConfiguration.AdDecisionServerUrl != nil || res.LivePreRollConfiguration.MaxDurationSeconds != nil {
		temp := map[string]interface{}{}
		if res.LivePreRollConfiguration.AdDecisionServerUrl != nil {
			temp["ad_decision_server_url"] = res.LivePreRollConfiguration.AdDecisionServerUrl
		}
		if res.LivePreRollConfiguration.MaxDurationSeconds != nil {
			temp["max_duration_seconds"] = res.LivePreRollConfiguration.MaxDurationSeconds
		}
		d.Set("live_pre_roll_configuration", []interface{}{temp})
	}

	if res.LogConfiguration != nil {
		if res.LogConfiguration.PercentEnabled != nil {
			d.Set("log_configuration", []interface{}{map[string]interface{}{
				"percent_enabled": res.LogConfiguration.PercentEnabled,
			}})
		}
	} else {
		d.Set("log_configuration", []interface{}{map[string]interface{}{
			"percent_enabled": 0,
		}})
	}

	if aws.BoolValue(res.ManifestProcessingRules.AdMarkerPassthrough.Enabled) {
		d.Set("manifest_processing_rules", []interface{}{map[string]interface{}{
			"ad_marker_passthrough": []interface{}{map[string]interface{}{
				"enabled": res.ManifestProcessingRules.AdMarkerPassthrough.Enabled,
			}},
		}})
	}

	d.Set("name", res.Name)
	d.Set("personalization_threshold_seconds", res.PersonalizationThresholdSeconds)
	d.Set("playback_configuration_arn", res.PlaybackConfigurationArn)
	d.Set("playback_endpoint_prefix", res.PlaybackEndpointPrefix)
	d.Set("session_initialization_endpoint_prefix", res.SessionInitializationEndpointPrefix)
	d.Set("slate_ad_url", res.SlateAdUrl)

	arn := aws.StringValue(res.PlaybackConfigurationArn)
	tags, err := ListTags(conn, arn)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error listing tags for media store container (%s): %s", arn, err))
	}
	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)
	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %w", err))
	}
	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags_all: %w", err))
	}
	d.Set("transcode_profile_name", res.TranscodeProfileName)
	d.Set("video_content_source_url", res.VideoContentSourceUrl)

	return nil
}

func resourcePlaybackConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MediaTailorConn

	_, err := conn.DeletePlaybackConfiguration(&mediatailor.DeletePlaybackConfigurationInput{Name: aws.String(d.Get("name").(string))})
	if err != nil {
		return diag.FromErr(fmt.Errorf("error while deleting the resource: %v", err))
	}

	return nil
}

func getPlaybackConfigurationInput(d *schema.ResourceData) mediatailor.PutPlaybackConfigurationInput {
	var params mediatailor.PutPlaybackConfigurationInput

	if v, ok := d.GetOk("ad_decision_server_url"); ok {
		params.AdDecisionServerUrl = aws.String(v.(string))
	}
	params.AvailSuppression = &mediatailor.AvailSuppression{}
	if v, ok := d.GetOk("avail_suppression_mode"); ok && v != nil {
		params.AvailSuppression.Mode = aws.String(v.(string))
	}
	if v, ok := d.GetOk("avail_suppression_value"); ok && v != nil {
		params.AvailSuppression.Value = aws.String(v.(string))
	}
	if v, ok := d.GetOk("bumper"); ok && v.([]interface{})[0] != nil {
		val := v.([]interface{})[0].(map[string]interface{})
		temp := mediatailor.Bumper{}
		if str, ok := val["end_url"]; ok {
			temp.EndUrl = aws.String(str.(string))
		}
		if str, ok := val["start_url"]; ok {
			temp.StartUrl = aws.String(str.(string))
		}
		params.Bumper = &temp
	}
	if v, ok := d.GetOk("cdn_configuration"); ok && v.([]interface{})[0] != nil {
		val := v.([]interface{})[0].(map[string]interface{})
		temp := mediatailor.CdnConfiguration{}
		if str, ok := val["ad_segment_url_prefix"]; ok {
			temp.AdSegmentUrlPrefix = aws.String(str.(string))
		}
		if str, ok := val["content_segment_url_prefix"]; ok {
			temp.ContentSegmentUrlPrefix = aws.String(str.(string))
		}
		params.CdnConfiguration = &temp
	}
	params.DashConfiguration = &mediatailor.DashConfigurationForPut{}
	if v, ok := d.GetOk("dash_mpd_location"); ok && v != nil {
		params.DashConfiguration.MpdLocation = aws.String(v.(string))
	}
	if v, ok := d.GetOk("dash_origin_manifest_type"); ok && v != nil {
		params.DashConfiguration.OriginManifestType = aws.String(v.(string))
	}
	if v, ok := d.GetOk("live_pre_roll_configuration"); ok && v.([]interface{})[0] != nil {
		val := v.([]interface{})[0].(map[string]interface{})
		temp := mediatailor.LivePreRollConfiguration{}
		if str, ok := val["ad_decision_server_url"]; ok {
			temp.AdDecisionServerUrl = aws.String(str.(string))
		}
		if integer, ok := val["max_duration_seconds"]; ok {
			temp.MaxDurationSeconds = aws.Int64(int64(integer.(int)))
		}
		params.LivePreRollConfiguration = &temp
	}
	if v, ok := d.GetOk("manifest_processing_rules"); ok && v.([]interface{})[0] != nil {
		val := v.([]interface{})[0].(map[string]interface{})
		temp := mediatailor.ManifestProcessingRules{}
		if v2, ok := val["ad_marker_passthrough"]; ok {
			temp2 := mediatailor.AdMarkerPassthrough{}
			val2 := v2.([]interface{})[0].(map[string]interface{})
			if boolean, ok := val2["enabled"]; ok {
				temp2.Enabled = aws.Bool(boolean.(bool))
			}
			temp.AdMarkerPassthrough = &temp2
		}
		params.ManifestProcessingRules = &temp
	}
	if v, ok := d.GetOk("name"); ok {
		params.Name = aws.String(v.(string))
	}
	if v, ok := d.GetOk("personalization_threshold_seconds"); ok {
		params.PersonalizationThresholdSeconds = aws.Int64(int64(v.(int)))
	}
	if v, ok := d.GetOk("slate_ad_url"); ok {
		params.SlateAdUrl = aws.String(v.(string))
	}
	if v, ok := d.GetOk("transcode_profile_name"); ok {
		params.TranscodeProfileName = aws.String(v.(string))
	}
	if v, ok := d.GetOk("video_content_source_url"); ok {
		params.VideoContentSourceUrl = aws.String(v.(string))
	}
	return params
}
