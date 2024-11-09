FROM python@sha256:a31cbb4db18c6f09e3300fa85b77f6d56702501fcb9bdb8792ec702a39ba6200

COPY . .

RUN pip install -r requirements.txt
EXPOSE 8001:8001

CMD ["fastapi", "run", "--port", "8001"]
