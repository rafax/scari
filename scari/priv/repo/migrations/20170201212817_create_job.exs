defmodule Scari.Repo.Migrations.CreateJob do
  use Ecto.Migration

  def change do
    create table(:jobs) do
      add :id, :binary, primary_key: true
      add :output, :string
      add :source, :string
      add :status, :string
      add :storage_id, :string, :null 
      add :lease_id, :string, :null

      timestamps()
    end

  end
end
