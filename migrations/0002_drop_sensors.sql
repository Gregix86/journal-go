-- Le suivi de capteurs (ESP32/MQTT) a ete retire du site : ces tables ne
-- sont plus utilisees par l'application. On les supprime proprement.

DROP TABLE IF EXISTS sensor_readings;
DROP TABLE IF EXISTS devices;
