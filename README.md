# Fluent Bit plugin for Yandex Cloud Logging

[Fluent Bit](https://fluentbit.io)
[output](https://docs.fluentbit.io/manual/concepts/data-pipeline/output)
for
[Yandex Cloud Logging](https://cloud.yandex.ru/docs/logging).

## Configuration parameters

All configuration parameters can be templated via [metadata service](#metadata). You can define templated value as follows: `{{key:default}}` or `{{key}}` (in this case `""` used as default value). All strings matched this template will be replaced by metadata values. In `default_payload` parameter you also can use this template as a JSON value (without quotes), so that JSON struct or array will be included in default payload.

| Key | Description | 
|:---|:---|
| `group_id`        | (_optional_) [Log group](https://cloud.yandex.ru/docs/logging/concepts/log-group) ID. Has higher priority than `folder_id`. |
| `folder_id`       | (_optional_) Folder ID. Has lower priority than `group_id`. Can be auto-detected via [metadata service](#metadata) if `group_id` and `folder_id` are not set. |
| `resource_type`   | (_optional_) Resource type of log entries. Can be templated via entry payload as follows: `{entry/json/path}`. | 
| `resource_id`     | (_optional_) Resource id of log entries. Can be templated via entry payload as follows: `{entry/json/path}`. | 
| `message_tag_key` | Key of the field to be assigned to the message tag. By default, will be skipped. | 
| `message_key`     | Key of the field, which will go to `message` attribute of LogEntry. | 
| `level_key`       | Key of the field, which contains log level, optional. |
| `default_level`   | (_optional_) Default level for messages, i.e., `INFO`. |
| `default_payload` | (_optional_) String with default JSON payload for entries (will be merged together with custom entry payload). |
| `authorization`   | See [Authorization](#authorization) section below. |

### Metadata

[Metadata service documentation](https://cloud.yandex.com/en/docs/compute/concepts/vm-metadata).

Metadata url can be configured through environment variable `YC_METADATA_URL`. By default it's `http://169.254.169.254`.

### Authorization

Configuration parameter `authorization` may have one of the following values:

| Value | Description |
|:---|:---|
|`instance-service-account` | run on behalf of instance service account |
| `iam-token` | environment variable `YC_TOKEN` <br> must contain a valid IAM token for authorization |
| `iam-key-file:/path/key.json` | use IAM key for authorization |

To create the key file, use [yc cli](https://cloud.yandex.ru/docs/cli/cli-ref/managed-services/iam/key/create).
Example:
```bash
  yc iam key create --service-account-name my-service-account --output key.json
```

### Configuration example

```
[OUTPUT]
    Name            yc-logging
    Match           *
    group_id        abc_{{group-id}}
    resource_type   {{resource}}_{resource/type}
    resource_id     {resource/id}
    message_key     text
    level_key       severity
    default_level   WARN
    default_payload {"num":5, "str": "string", "bool": true, "host":"{{instance/hostname}}", "not-found":"{{not/found:default}}", "struct": {{instance/disks/0}}, "array": {{instance/disks}} }
    authorization   instance-service-account
```
