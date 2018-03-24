FROM alpine
ADD homeinsight homeinsight
EXPOSE 8080
CMD [ "./homeinsight" ]