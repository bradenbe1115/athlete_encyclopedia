FROM python:3.11-slim

ENV ENC_HOME=/athlete_encyclopedia

WORKDIR /app
COPY ./back_end/ingestion/get_espn_data get_espn_data

WORKDIR get_espn_data
RUN pip install --upgrade pip \
    && pip install -r requirements.lock.txt

ENTRYPOINT ["python", "-m"]
