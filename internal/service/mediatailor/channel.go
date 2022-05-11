package mediatailor

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mediatailor"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func ResourceChannel() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceChannelCreate,
		ReadContext:   resourceChannelRead,
		UpdateContext: resourceChannelUpdate,
		DeleteContext: resourceChannelDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"channel_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"channel_state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creation_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"filler_slate": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"source_location_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"vod_source_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"last_modified_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"outputs": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"dash_manifest_windows_seconds": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(30, 3600),
						},
						"dash_min_buffer_time_seconds": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(2, 60),
						},
						"dash_min_update_period_seconds": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(2, 60),
						},
						"dash_suggested_presentation_delay_seconds": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(2, 60),
						},
						"hls_manifest_windows_seconds": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(30, 3600),
						},
						"manifest_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"playback_url": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"source_group": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"playback_mode": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"LINEAR", "LOOP"}, false),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"tier": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"BASIC", "STANDARD"}, false),
			},
		},
	}
}

func resourceChannelCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MediaTailorConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	var params = getChannelInput(d)

	if len(tags) > 0 {
		params.Tags = Tags(tags.IgnoreAWS())
	}

	channel, err := conn.CreateChannel(&params)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error while creating the channel: %v", err))
	}
	d.SetId(aws.StringValue(channel.Arn))

	return resourceChannelRead(ctx, d, meta)
}

func resourceChannelRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceChannelUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceChannelDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MediaTailorConn

	_, err := conn.DeleteChannel(&mediatailor.DeleteChannelInput{ChannelName: aws.String(d.Get("channel_name").(string))})
	if err != nil {
		return diag.FromErr(fmt.Errorf("error while deleting the resource: %v", err))
	}

	return nil
}

func getChannelInput(d *schema.ResourceData) mediatailor.CreateChannelInput {
	var params mediatailor.CreateChannelInput

	if v, ok := d.GetOk("channel_name"); ok {
		params.ChannelName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("filler_slate"); ok && v.([]interface{})[0] != nil {
		val := v.([]interface{})[0].(map[string]interface{})
		temp := mediatailor.SlateSource{}
		if str, ok := val["source_location_name"]; ok {
			temp.SourceLocationName = aws.String(str.(string))
		}
		if str, ok := val["vod_source_name"]; ok {
			temp.VodSourceName = aws.String(str.(string))
		}
		params.FillerSlate = &temp
	}

	if v, ok := d.GetOk("outputs"); ok && v.([]interface{})[0] != nil {
		outputs := v.([]interface{})

		var res []*mediatailor.RequestOutputItem

		for _, output := range outputs {
			current := output.(map[string]interface{})
			temp := mediatailor.RequestOutputItem{}

			if str, ok := current["manifest_name"]; ok {
				temp.ManifestName = aws.String(str.(string))
			}
			if str, ok := current["source_group"]; ok {
				temp.SourceGroup = aws.String(str.(string))
			}

			if num, ok := current["hls_manifest_windows_seconds"]; ok && num.(int) != 0 {
				tempHls := mediatailor.HlsPlaylistSettings{}
				tempHls.ManifestWindowSeconds = aws.Int64(int64(num.(int)))
				temp.HlsPlaylistSettings = &tempHls
			}

			tempDash := mediatailor.DashPlaylistSettings{}
			if num, ok := current["dash_manifest_windows_seconds"]; ok && num.(int) != 0 {
				tempDash.ManifestWindowSeconds = aws.Int64(int64(num.(int)))
			}
			if num, ok := current["dash_min_buffer_time_seconds"]; ok && num.(int) != 0 {
				tempDash.MinBufferTimeSeconds = aws.Int64(int64(num.(int)))
			}
			if num, ok := current["dash_min_update_period_seconds"]; ok && num.(int) != 0 {
				tempDash.MinBufferTimeSeconds = aws.Int64(int64(num.(int)))
			}
			if num, ok := current["dash_suggested_presentation_delay_seconds"]; ok && num.(int) != 0 {
				tempDash.SuggestedPresentationDelaySeconds = aws.Int64(int64(num.(int)))
			}
			if tempDash != (mediatailor.DashPlaylistSettings{}) {
				temp.DashPlaylistSettings = &tempDash
			}

			res = append(res, &temp)
		}
		params.Outputs = res
	}

	if v, ok := d.GetOk("playback_mode"); ok {
		params.PlaybackMode = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tier"); ok {
		params.Tier = aws.String(v.(string))
	}

	return params
}
