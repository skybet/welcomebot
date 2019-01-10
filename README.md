# WelcomeBot

Want to welcome people to your Slack?

> Oh, hi! You're new to the company.
> Welcome to the team!
>
> Here's some channels you might want to join:
> 
> - #help-with-stuff
> - #cute-animals
> - #just-office-things

How about when someone joins a specific channel?

> Hello!
> Thanks for joining the FancyTeam channel.
>
> We look after Foo, Bar, and Baz services.
>
> We use Monkeybot in this channel to triage issues, so attact their attention with `@monkeybot help`

Example configs in `config.json`

## Building and usage

You will need a Slack token in order to run this.  Welcomebot expects this to be provided as the environment variable `SLACK_TOKEN`.  You will also need to provide a config.json file in the current directory.  See the example config.json in this repo for what should go into that file.

The easiest way of building and running this is in Docker.  First of all build a Docker container:

```
docker build -t welcomebot .
```

And then you can run it as follows:

```
docker run -v config.json -e SLACK_TOKEN=foo welcomebot
```
