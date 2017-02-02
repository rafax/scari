defmodule Scari.JobView do
  use Scari.Web, :view

  def render("index.json", %{jobs: jobs}) do
    %{data: render_many(jobs, Scari.JobView, "job.json")}
  end

  def render("show.json", %{job: job}) do
    %{data: render_one(job, Scari.JobView, "job.json")}
  end

  def render("job.json", %{job: job}) do
    %{id: job.id,
      output: job.output,
      source: job.source,
      status: job.status,
      storage_id: job.storage_id,
      lease_id: job.lease_id}
  end
end
