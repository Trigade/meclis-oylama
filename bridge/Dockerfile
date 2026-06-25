FROM python:3.12-slim
WORKDIR /app
COPY pyproject.toml .
COPY bridge/ ./bridge/
CMD ["python", "-m", "bridge.main"]