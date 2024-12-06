FROM python@sha256:a31cbb4db18c6f09e3300fa85b77f6d56702501fcb9bdb8792ec702a39ba6200
ARG USER_UID=1001
RUN adduser --shell /bin/sh --system --group --uid "${USER_UID}" default
COPY . .

RUN pip install -r requirements.txt
EXPOSE 8000:8000

CMD ["fastapi", "run", "--port", "8000"]
