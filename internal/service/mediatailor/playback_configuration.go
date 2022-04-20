package mediatailor

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/mediatailor"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"regexp"
)

func ResourcePlaybackConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePlaybackConfigurationCreate,
		ReadContext:   resourcePlaybackConfigurationRead,
		UpdateContext: resourcePlaybackConfigurationUpdate,
		DeleteContext: resourcePlaybackConfigurationDelete,
		Schema: map[string]*schema.Schema{
			"ad_decision_server_url": {
				Type:     schema.TypeString,
				Required: true,
			},
			"avail_suppression": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"mode": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"OFF", "BEHIND_LIVE_EDGE"}, false),
						},
						"value": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringMatch(regexp.MustCompile(`^\d{2}:\d{2}:\d{2}$`), "must be valid HH:MM:SS string"),
						},
					},
				},
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
			"configuration_aliases": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type:     schema.TypeMap,
					Optional: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
			},
			"dash_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"manifest_endpoint_prefix": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"mpd_location": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"DISABLED", "EMT_DEFAULT"}, false),
						},
						"origin_manifest_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"SINGLE_PERIOD", "MULTI_PERIOD"}, false),
						},
					},
				},
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
							Type:     schema.TypeInt,
							Optional: true,
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
			"tags": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"transcode_profile_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"video_content_source_url": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 512),
			},
			"last_updated": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourcePlaybackConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MediaTailorConn

	var params mediatailor.PutPlaybackConfigurationInput

	if v, ok := d.GetOk("ad_decision_server_url"); ok {
		params.AdDecisionServerUrl = aws.String(v.(string))
	}
	if v, ok := d.GetOk("avail_suppression"); ok && v.([]interface{})[0] != nil {
		val := v.([]interface{})[0].(map[string]interface{})
		temp := mediatailor.AvailSuppression{}
		if str, ok := val["mode"]; ok {
			temp.Mode = aws.String(str.(string))
		}
		if str, ok := val["value"]; ok {
			temp.Value = aws.String(str.(string))
		}
		params.AvailSuppression = &temp
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
	if v, ok := d.GetOk("configuration_aliases"); ok {
		val := v.(map[string]map[string]*string)
		params.ConfigurationAliases = val
	}
	if v, ok := d.GetOk("dash_configuration"); ok && v.([]interface{})[0] != nil {
		val := v.([]interface{})[0].(map[string]interface{})
		temp := mediatailor.DashConfigurationForPut{}
		if str, ok := val["mpd_location"]; ok {
			temp.MpdLocation = aws.String(str.(string))
		}
		if str, ok := val["origin_manifest_type"]; ok {
			temp.OriginManifestType = aws.String(str.(string))
		}
		params.DashConfiguration = &temp
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
	tempMap := make(map[string]*string)
	if v, ok := d.GetOk("tags"); ok {
		val := v.(map[string]interface{})
		for k, value := range val {
			temp := value.(string)
			tempMap[k] = &temp
		}
	}
	params.Tags = tempMap
	if v, ok := d.GetOk("transcode_profile_name"); ok {
		params.TranscodeProfileName = aws.String(v.(string))
	}
	if v, ok := d.GetOk("video_content_source_url"); ok {
		params.VideoContentSourceUrl = aws.String(v.(string))
	}

	playbackConfiguration, err := conn.PutPlaybackConfiguration(&params)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error while creating the playback configuration: %v", err))
	}
	d.SetId(*playbackConfiguration.PlaybackConfigurationArn)
	return resourcePlaybackConfigurationRead(ctx, d, meta)
}

func resourcePlaybackConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MediaTailorConn

	resourceName := d.Get("name").(string)
	if len(resourceName) == 0 && len(d.Id()) > 0 {
		resourceArn, err := arn.Parse(d.Id())
		if err != nil {
			return diag.FromErr(fmt.Errorf("error parsing the name from resource arn: %v", err))
		}
		resourceName = resourceArn.Resource
	}

	res, err := conn.GetPlaybackConfiguration(&mediatailor.GetPlaybackConfigurationInput{Name: aws.String(resourceName)})
	if err != nil {
		return diag.FromErr(fmt.Errorf("error while retrieving the resource: %v", err))
	}

	d.Set("ad_decision_server_url", res.AdDecisionServerUrl)
	if res.AvailSuppression != nil {
		temp := map[string]interface{}{}
		if res.AvailSuppression.Mode != nil {
			temp["mode"] = res.AvailSuppression.Mode
		}
		if res.AvailSuppression.Value != nil {
			temp["value"] = res.AvailSuppression.Value
		}
		d.Set("avail_suppression", []interface{}{temp})
	}
	if res.Bumper != nil {
		temp := map[string]interface{}{}
		if res.Bumper.StartUrl != nil {
			temp["end_url"] = res.Bumper.EndUrl
		}
		if res.Bumper.StartUrl != nil {
			temp["start_url"] = res.Bumper.StartUrl
		}
		d.Set("bumper", []interface{}{temp})
	}
	if res.CdnConfiguration != nil {
		temp := map[string]interface{}{}
		if res.CdnConfiguration.AdSegmentUrlPrefix != nil {
			temp["ad_segment_url_prefix"] = res.CdnConfiguration.AdSegmentUrlPrefix
		}
		if res.CdnConfiguration.ContentSegmentUrlPrefix != nil {
			temp["content_segment_url_prefix"] = res.CdnConfiguration.ContentSegmentUrlPrefix
		}
		d.Set("cdn_configuration", []interface{}{temp})
	}
	d.Set("configuration_aliases", res.ConfigurationAliases)
	if res.DashConfiguration != nil {
		temp := map[string]interface{}{}
		if res.DashConfiguration.ManifestEndpointPrefix != nil {
			temp["manifest_endpoint_prefix"] = res.DashConfiguration.ManifestEndpointPrefix
		}
		if res.DashConfiguration.MpdLocation != nil {
			temp["mpd_location"] = res.DashConfiguration.MpdLocation
		}
		if res.DashConfiguration.OriginManifestType != nil {
			temp["origin_manifest_type"] = res.DashConfiguration.OriginManifestType
		}
		d.Set("dash_configuration", []interface{}{temp})
	}
	if res.HlsConfiguration != nil {
		temp := map[string]interface{}{}
		if res.HlsConfiguration.ManifestEndpointPrefix != nil {
			temp["manifest_endpoint_prefix"] = res.HlsConfiguration.ManifestEndpointPrefix
		}
		d.Set("hls_configuration", []interface{}{temp})
	}
	if res.LivePreRollConfiguration != nil {
		temp := map[string]interface{}{}
		if res.LivePreRollConfiguration.AdDecisionServerUrl != nil {
			temp["ad_decision_server_url"] = res.LivePreRollConfiguration.AdDecisionServerUrl
		}
		if res.LivePreRollConfiguration.MaxDurationSeconds != nil {
			temp["max_duration_seconds"] = res.LivePreRollConfiguration.MaxDurationSeconds
		}
		d.Set("live_pre_roll_configuration", temp)
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
	if *res.ManifestProcessingRules.AdMarkerPassthrough.Enabled == true {
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
	d.Set("tags", res.Tags)
	d.Set("transcode_profile_name", res.TranscodeProfileName)
	d.Set("video_content_source_url", res.VideoContentSourceUrl)

	return nil
}

func resourcePlaybackConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourcePlaybackConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}