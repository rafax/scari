defmodule Scari.JobControllerTest do
  use Scari.ConnCase

  alias Scari.Job
  @valid_attrs %{lease_id: "some content", output: "some content", source: "some content", status: "some content", storage_id: "some content"}
  @invalid_attrs %{}

  setup %{conn: conn} do
    {:ok, conn: put_req_header(conn, "accept", "application/json")}
  end

  test "lists all entries on index", %{conn: conn} do
    conn = get conn, job_path(conn, :index)
    assert json_response(conn, 200)["data"] == []
  end

  test "shows chosen resource", %{conn: conn} do
    job = Repo.insert! %Job{}
    conn = get conn, job_path(conn, :show, job)
    assert json_response(conn, 200)["data"] == %{"id" => job.id,
      "output" => job.output,
      "source" => job.source,
      "status" => job.status,
      "storage_id" => job.storage_id,
      "lease_id" => job.lease_id}
  end

  test "renders page not found when id is nonexistent", %{conn: conn} do
    assert_error_sent 404, fn ->
      get conn, job_path(conn, :show, -1)
    end
  end

  test "creates and renders resource when data is valid", %{conn: conn} do
    conn = post conn, job_path(conn, :create), job: @valid_attrs
    assert json_response(conn, 201)["data"]["id"]
    assert Repo.get_by(Job, @valid_attrs)
  end

  test "does not create resource and renders errors when data is invalid", %{conn: conn} do
    conn = post conn, job_path(conn, :create), job: @invalid_attrs
    assert json_response(conn, 422)["errors"] != %{}
  end

  test "updates and renders chosen resource when data is valid", %{conn: conn} do
    job = Repo.insert! %Job{}
    conn = put conn, job_path(conn, :update, job), job: @valid_attrs
    assert json_response(conn, 200)["data"]["id"]
    assert Repo.get_by(Job, @valid_attrs)
  end

  test "does not update chosen resource and renders errors when data is invalid", %{conn: conn} do
    job = Repo.insert! %Job{}
    conn = put conn, job_path(conn, :update, job), job: @invalid_attrs
    assert json_response(conn, 422)["errors"] != %{}
  end

  test "deletes chosen resource", %{conn: conn} do
    job = Repo.insert! %Job{}
    conn = delete conn, job_path(conn, :delete, job)
    assert response(conn, 204)
    refute Repo.get(Job, job.id)
  end
end
