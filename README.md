# Filebeat

Filebeat harvest file logs by line and send to output. Only Linux is supported.

**Features**:

- Harvest from last offset.
- Automatically discover new log files.
- Harvest until file is deleted or moved.
- Donot depend on file path. If a file has been moved or recreated, harvest will not be affected.

### Config

Config is a json file.

```json
{
    "input": [
        "/var/log/nginx/*.log"
    ],
    "output": {
        "type": "udp",
        "settings": {
            "address": ""
        }
    },
    "filter": {
        "type": "regex",
        "settings": {
            "pattern": ""
        }
    },
    "harvest_inteval": 1,
    "dump_inteval": 30,
    "discover": true
}
```

- input: Input file paths, support glob.
- output: Support only udp address now.
- filter: Mismatched lines are discarded.
    - regex: match by regex.
    ```json
    "filter": {
        "type": "regex",
        "settings": {
            "pattern": ""
        }
    }
    ```
    - in/notin: match by word in/notin line.
    ```json
    "filter": {
        "type": "in",
        "settings": {
            "word": ""
        }
    }
    ```
- harvest_inteval: Harvest interval, default 1 seconds.
- dump_inteval: Interval of dump offset to file, default 30 seconds.
- discover: Auto discover log files matched by glob, default true.

### Usage

```
-c string
    config file (default "config.json")
```
