FROM alpine

WORKDIR /app
COPY  goRestAPI ./mpesa
RUN chmod +x ./mpesa
CMD [ "./mpesa" ]