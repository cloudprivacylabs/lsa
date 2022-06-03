FROM python:3.8-slim-buster
WORKDIR /app
RUN pip install --upgrade pip
RUN pip install  pyyaml psycopg[binary]
COPY *.py /app
CMD python3 server.py

