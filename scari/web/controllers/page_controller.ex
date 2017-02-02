defmodule Scari.PageController do
  use Scari.Web, :controller

  def index(conn, _params) do
    render conn, "index.html"
  end
end
