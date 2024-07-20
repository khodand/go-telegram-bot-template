# Telegram Bot Template

This is a template for creating a Telegram bot using Go. It uses the [telebot](https://github.com/tucnak/telebot) library.

## Suggestions and Improvements

We welcome your suggestions and advice! Please feel free to submit issues if you have any ideas on how to improve this template.

## Usage

After cloning the repository:

1. Replace token in `cmd/config/secret.yaml` with your bot token.
2. Uncomment `# secret.*` in .gitignore to hide your token. And run `git rm --cached cmd/config/secret.yaml`.
3. Replace module name in `go.mod` with your module name.
It will break the import path in the code, so you need to replace all import paths in the code with the new module name.
4. Replace local prefixes in `golangci.yml` with your module name.
5. Set ENV variable `CONFIG_PATH=cmd/config/secret.yaml`

### Docker and github workflows

This template includes a Dockerfile and a github workflow to build and run the bot in a container.

You have to create github secrets in your repository with following names or replace this name in `.github/workflows/deploy.yml`:
- `TIMEWEB_FULL_HOST` - full host for your server (e.g. `user@1.1.1.1`)
- `TIMEWEB_SSH_KEY` - private SSH key for your server
- `TIMEWEB_HOST` - host for your server (e.g. `1.1.1.1`)
- `DOCKERHUB_USERNAME` - your dockerhub username
- `DOCKERHUB_TOKEN` - your dockerhub secret token
