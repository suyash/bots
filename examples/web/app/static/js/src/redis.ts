import debug from "debug";

import Chat from "@suy/bots-web-client";
import { Message } from "@suy/bots-web-client/lib/message";

debug.enable("*");

const log: debug.IDebugger = debug("app");

window.addEventListener("DOMContentLoaded", loaded);

let userMessage: HTMLTemplateElement = null;
let botMessage: HTMLTemplateElement = null;

async function loaded(): Promise<void> {
    const chat: Chat = new Chat(`ws://${window.location.host}/echo_redis_chat`);
    await chat.open();
    log("open");

    const container: HTMLDivElement = document.querySelector(".messages");

    userMessage = document.querySelector("#user-message");
    botMessage = document.querySelector("#bot-message");

    const form: HTMLFormElement = document.querySelector(".form");
    const message: HTMLInputElement = document.querySelector("#message");
    form.addEventListener("submit", (e: Event): void => {
        e.preventDefault();

        log("sending", message.value);
        chat.say(message.value);

        message.value = "";
    });

    for await (const m of chat.messages()) {
        const msg: Message = m;
        log("received", msg);

        let node: DocumentFragment = newBotMessage(msg);
        if (msg.source === "user") {
            node = newUserMessage(msg);
        }

        let pnode: HTMLDivElement = document.querySelector(`.message[data-next-ts="${msg.id}"]`);
        if (!pnode && msg.prev) {
            pnode = document.querySelector(`.message[data-ts="${msg.prev.id}"]`);
        }

        let nnode: HTMLDivElement = document.querySelector(`.message[data-prev-ts="${msg.id}"]`);
        if (!nnode && msg.next) {
            nnode = document.querySelector(`.message[data-ts="${msg.next.id}"]`);
        }

        if (msg.prev && pnode && pnode.nextSibling) {
            pnode.parentElement.insertBefore(node, pnode.nextSibling);
        } else if (msg.next && nnode) {
            nnode.parentElement.insertBefore(node, nnode);
        } else {
            // first
            container.appendChild(node);
        }
    }

    log("closed");
}

function newUserMessage(msg: Message): DocumentFragment {
    userMessage.content.querySelector("section").textContent = msg.text;

    userMessage.content.querySelector(".message").setAttribute("data-ts", msg.id.toString());

    if (msg.prev) {
        userMessage.content.querySelector(".message").setAttribute("data-prev-ts", msg.prev.id.toString());
    }

    if (msg.next) {
        userMessage.content.querySelector(".message").setAttribute("data-next-ts", msg.next.id.toString());
    }

    return document.importNode(userMessage.content, true);
}

function newBotMessage(msg: Message): DocumentFragment {
    botMessage.content.querySelector("section").textContent = msg.text;
    botMessage.content.querySelector(".message").setAttribute("data-ts", msg.id.toString());
    return document.importNode(botMessage.content, true);
}
