from python:3.8
workdir /app
copy debug.py .
entrypoint ["python3.8", "debug.py"]
