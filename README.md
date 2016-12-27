Build image and tag as scari-worker:
docker build . -t=scari-worker

Download to mp3:

`docker run -v /Users/rafal/dev/scari/out:/out --rm scari-worker  youtube-dl -o '/out/%(title)s.%(ext)s' -x --audio-format mp3 https://www.youtube.com/watch\?v\=Ee1v_SuECRk\&feature\=youtu.be`