FROM python:3-alpine

WORKDIR /app

COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt

COPY ./app /app

ENTRYPOINT ["python"]
CMD ["app.py"]

