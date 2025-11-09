FROM python:3.9-slim-buster

COPY requirements.txt /tmp/
RUN pip install -r /tmp/requirements.txt

RUN mkdir -p /src
COPY back_end/src /src/

WORKDIR /src

RUN pip install -e .

# bootstrap athletes db
CMD ["python", "athletes/bootstrap_db.py"]