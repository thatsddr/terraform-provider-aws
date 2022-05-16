---
subcategory: "Elemental MediaTailor"
layout: "aws"
page_title: "AWS: aws_media_tailor_playback_configuration"
description: |-
Manages a MediaTailor Playback Configuration
---

# Resource: aws_media_tailor_channel

Provides an Elemental MediaTailor Channel.

## Example Usage

```terraform
resource "aws_media_tailor_channel" "example" {
  channel_name = "example-channel"
  outputs {
    manifest_name                             = "default"
    source_group                              = "default"
    hls_manifest_windows_seconds              = 30
  }
  playback_mode = "LOOP"
  tier = "BASIC"
}
```

## Argument Reference
The following arguments are supported:

* `channel_name` - (Required) The name of the channel.
* `playback_mode` - (Required) The type of playback mode for this channel. Can be either LINEAR or LOOP.
* `source_group` - (Required) A string used to match which HttpPackageConfiguration is used for each VodSource.
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `tier` - (Required)  The tier for this channel. STANDARD tier channels can contain live programs.

### `filler_slate`
The slate used to fill gaps between programs in the schedule. You must configure filler slate if your channel uses the LINEAR PlaybackMode.

* `source_location_name` - (Optional) The name of the source location where the slate VOD source is stored.
* `vod_source_name` - (Optional) The slate VOD source name. The VOD source must already exist in a source location before it can be used for slate.

## `outputs`
The channel's output properties.

* `dash_manifest_windows_seconds` - (Optional) The total duration (in seconds) of each dash manifest.
* `dash_min_buffer_time_seconds` - (Optional) Minimum amount of content (measured in seconds) that a player must keep available in the buffer.
* `dash_min_update_period_seconds` - (Optional) Minimum amount of time (in seconds) that the player should wait before requesting updates to the manifest.
* `dash_suggested_presentation_delay_seconds` - (Optional) Amount of time (in seconds) that the player should be from the live point at the end of the manifest.
* `hls_manifest_windows_seconds` - (Optional) The total duration (in seconds) of each hls manifest.
* `manifest_name` - (Required) The name of the manifest for the channel. The name appears in the PlaybackUrl.

## Attributes Reference
In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the channel.
* `channel_state` - Returns whether the channel is running or not.
* `creation_time` - The timestamp of when the channel was created.
* `last_modified_time` - The timestamp of when the channel was last modified.

## `outputs`

* `playback_url` - The URL used for playback by content players.

## Import

Channels can be imported using their ARN as identifier. For example:

```
  $ terraform import aws_media_tailor_channel.example arn:aws:mediatailor:us-east-1:000000000000:channel/example
```