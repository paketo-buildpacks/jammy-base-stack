package io.paketo;

import java.io.IOException;
import java.io.OutputStream;
import java.net.InetSocketAddress;

import com.sun.net.httpserver.HttpServer;
import com.sun.net.httpserver.HttpHandler;
import com.sun.net.httpserver.HttpExchange;

public class Main {
  public static void main(String[] args) throws Exception {
    final int port = Integer.parseInt(System.getenv("PORT"));

    HttpServer server = HttpServer.create(new InetSocketAddress(port), 0);

    server.createContext("/", new HttpHandler(){
        @Override
        public void handle(HttpExchange ex) throws IOException {
            final String body = "Hello World! Java version " + Runtime.version().toString();

            OutputStream out = ex.getResponseBody();
            ex.sendResponseHeaders(200, body.length());
            out.write(body.getBytes());
            out.close();
        }
    });

    server.start();
    System.out.println("Started server on port: " + port);
  }
}
