# slog-handler-adapter

slog-handler-adapter is an implementation of slog.Handler for third-party logging libraries, even though it may not seem to make much sense

## Overview

This project aims to implement the following third-party logging libraries:

- [x] [logrus](https://github.com/sirupsen/logrus)
- [ ] [zap](https://github.com/uber-go/zap)
- [ ] [zerolog](https://github.com/rs/zerolog)

Here, we've encapsulated the implementations of these third-party logging libraries as slog.Handler, allowing us to embed them within slog.Logger. In this context, we've disabled the third-party "log level control" and "timestamp output" functionalities, moving them to the slog layer.

i acknowledge that the code may not be perfect, and welcome contributions and suggestions for improvement. If you have any ideas, bug reports, or would like to contribute in any way, please feel free to open an issue or a pull request. Your feedback and contributions are highly appreciated, and they help us make this project better.

## Conclusion

This repository is for sharing purposes only, and I do not plan to actively maintain or expand it. Feel free to use the code as needed, but please be aware that I may not provide ongoing support or updates.