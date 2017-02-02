# This file is responsible for configuring your application
# and its dependencies with the aid of the Mix.Config module.
#
# This configuration file is loaded before any dependency and
# is restricted to this project.
use Mix.Config

# General application configuration
config :scari,
  ecto_repos: [Scari.Repo]

# Configures the endpoint
config :scari, Scari.Endpoint,
  url: [host: "localhost"],
  secret_key_base: "ugzq8pvagtz6armCYTWLUygxoeWFxpKXdPVRvp5TVzQMemY7emLcFf4mUBUNdWce",
  render_errors: [view: Scari.ErrorView, accepts: ~w(html json)],
  pubsub: [name: Scari.PubSub,
           adapter: Phoenix.PubSub.PG2]

# Configures Elixir's Logger
config :logger, :console,
  format: "$time $metadata[$level] $message\n",
  metadata: [:request_id]

# Import environment specific config. This must remain at the bottom
# of this file so it overrides the configuration defined above.
import_config "#{Mix.env}.exs"
