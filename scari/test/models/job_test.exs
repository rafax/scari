defmodule Scari.JobTest do
  use Scari.ModelCase

  alias Scari.Job

  @valid_attrs %{lease_id: "some content", output: "some content", source: "some content", status: "some content", storage_id: "some content"}
  @invalid_attrs %{}

  test "changeset with valid attributes" do
    changeset = Job.changeset(%Job{}, @valid_attrs)
    assert changeset.valid?
  end

  test "changeset with invalid attributes" do
    changeset = Job.changeset(%Job{}, @invalid_attrs)
    refute changeset.valid?
  end
end
