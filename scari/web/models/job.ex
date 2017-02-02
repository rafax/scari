defmodule Scari.Job do
  use Scari.Web, :model

  schema "jobs" do
    field :output, :string
    field :source, :string
    field :status, :string
    field :storage_id, :string
    field :lease_id, :string

    timestamps()
  end

  @doc """
  Builds a changeset based on the `struct` and `params`.
  """
  def changeset(struct, params \\ %{}) do
    struct
    |> cast(params, [:output, :source, :status, :storage_id, :lease_id])
    |> validate_required([:output, :source, :status, :storage_id, :lease_id])
  end
end
