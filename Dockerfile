FROM alpine

WORKDIR /app
COPY  goRestAPI /app/mpesa
RUN chmod +x /app/mpesa
CMD [ "./mpesa" ]