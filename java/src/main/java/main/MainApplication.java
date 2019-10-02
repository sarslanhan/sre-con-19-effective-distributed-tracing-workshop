package main;

import cart.CartApiApplication;
import org.springframework.boot.builder.SpringApplicationBuilder;
import web.WebApplication;

public class MainApplication {
    public static void main(String[] args) {
        new SpringApplicationBuilder()
                .parent(MainConfig.class)
                .child(CartApiApplication.class)
                .sibling(WebApplication.class)
                .run(args);
    }
}
