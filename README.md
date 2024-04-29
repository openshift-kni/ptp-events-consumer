# ptp-events-consumer
  The ptp-events-consumer provide a consumer app example to use RH ptp events internal API. The example app peorically checks the ptp event publisher health (running in the ptp namespace), get ptp event state. The app also subscribes the ptp event notification, so that the ptp event pulblisher will post the ptp event state change to the consumer app.
  
  The consumer app deployment example with passed in node name, consumer app namespace and consumer app listening port could be found under the deployment directory. 
