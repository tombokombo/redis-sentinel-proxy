FROM alpine:3.10

MAINTAINER Tombo Kombo <tombo.hulata@gmail.com>

RUN apk add --update --no-cache python3 python3-dev py-virtualenv 

RUN mkdir /app

COPY testclient.py /app/testclient.py

COPY requirements.txt /app/requirements.txt

WORKDIR /app

RUN /usr/bin/virtualenv -p /usr/bin/python3 --system-site-packages /app/venv/ && \
    . /app/venv/bin/activate && pip install -r requirements.txt

ENV VIRTUAL_ENV=/app/venv
ENV PATH="$VIRTUAL_ENV/bin:$PATH"

CMD ["python","/app/testclient.py"]
