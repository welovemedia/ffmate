# ffmate: FFmpeg with a REST API, Web UI, Webhooks & Queues

**ffmate** bridges the gap between the raw power of FFmpeg and the need for user-friendly interfaces and robust integration capabilities. While FFmpeg is incredibly versatile, its complex command-line syntax can be daunting. **ffmate** solves this by providing simplified interfaces, ready-to-use presets, a robust queueing system, and advanced features like real-time webhook notifications and automated post-transcoding tasks

## Key Features

*   **User-Friendly Interfaces:** Offers a REST API, web interface and a modern web interface for easy interaction with FFmpeg.
  
*   **Ready-to-Use Presets:** Includes a comprehensive set of pre-configured transcoding presets for common use cases, simplifying common tasks.
    
*   **Robust Queueing System:** Manages transcoding tasks with a powerful queue that supports prioritization, grouping, and efficient processing.
    
*   **Webhook Notifications:** Enables real-time notifications to third-party systems about *all* transcoding events (start, progress, completion, errors), facilitating seamless integration into existing workflows and allowing for automated actions to be triggered based on transcoding status.
  
*   **Post-Transcoding Tasks:** Allows for the execution of custom scripts *immediately* after transcoding is complete.
    
*   **Output Filename Wildcard Support:** Enables dynamic filename generation using wildcards.

ffmates **featureset** (and roadmap) looks like this:

- [x] Rest API
    - [x] Tasks
    - [x] Webhooks
    - [x] Swagger documentation
- [x] Tasks
    - [x] Status (Queue, Running, Successful/Error/Canceled)
    - [x] Progress
    - [x] Wildcards
- [x] Queue
    - [x] Max concurrent tasks
- [x] Webhooks
    - [x] Create task
    - [x] Update task status
- [x] Update service    
- [ ] Presets
- [ ] Dashboard
- [ ] and more..

If you feel like the above description fits your needs, **welcome**! 

## Quick start guide

ffmate comes with a documented commandline.
Using **ffmate server** will start the server on the default port, 3000. This port can be overriden using the **-p** flag.

Use the API to add new Tasks. A Task includes the ffmpeg command to run together with the input and output file. \
One can make use of various wildcards:
- ${INPUT_FILE}
- ${OUTPUT_FILE}
- ${DATE_YEAR}
- ${DATE_SHORTYEAR}
- ${DATE_MONTH}
- ${DATE_DAY}
- ${TIME_HOUR}
- ${TIME_MINUTE}
- ${TIME_SECOND}
- ${TIMESTAMP_SECONDS}
- ${TIMESTAMP_MILLISECONDS}
- ${TIMESTAMP_MICROSECONDS}
- ${TIMESTAMP_NANOSECONDS}

## Swagger / Openapi

The swagger documentation can be found here **http://localhost:3000/swagger/index.html** once the server is running.

## Contributing

Thank you for considering contributing to ffmate! A contribution guide will be released in the future.

## License

ffmate is MIT Licensed. Copyright © 2025 by We love media
