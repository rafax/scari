defmodule HappyPathTest do
  use ExUnit.Case

  @apiServer  "http://localhost:3001/"
  @jobsUrl @apiServer <> "jobs"
  @source "https://www.youtube.com/watch?v=zbh0qhGmJ9U"
  @output "audio"

  test "created job should be returned by GET" do
    job_id = create_job()

    case HTTPoison.get(@jobsUrl<>"/"<>job_id) do
      {:ok, %HTTPoison.Response{status_code: 200, body: body}} ->
        job = Poison.Parser.parse!(body)["job"]
        assert job["id"]==job_id
    end
  end

  test "job should be leased after creation" do
    create_job()

    {job_id, lease_id} = case HTTPoison.post(@jobsUrl<>"/lease", "",[{"Content-Type", "text/json"}]) do
      {:ok, %HTTPoison.Response{status_code: 200, body: body}} ->
        lease = Poison.Parser.parse!(body)
        assert nil != lease["job"]["id"]
        assert nil != lease["leaseId"]
        {lease["job"]["id"], lease["leaseId"]}
    end

    complete_job_request =  Poison.encode! %{leaseId: lease_id, storageUrl: "https://storage.googleapis.com/scari-666.appspot.com/Steroids_In_CrossFit_-_Featuring_Johnny_Romano_WODdoc_P365_Episode_676.mp3"}

    {_,response} = HTTPoison.post(@jobsUrl<>"/"<>job_id<>"/complete" , complete_job_request,[{"Content-Type", "text/json"}]) 
    assert response.status_code == 200, response.body

  end

  defp create_job do
    request_body = Poison.encode! %{source: @source, outputType: @output}

    case HTTPoison.post(@jobsUrl,  request_body, [{"Content-Type", "text/json"}]) do
      {:ok, %HTTPoison.Response{status_code: 200, body: body}} ->
        job = Poison.Parser.parse!(body)["job"]
        assert job["output"] == @output
        assert job["source"] == @source
        job["id"]
    end
  end

end
