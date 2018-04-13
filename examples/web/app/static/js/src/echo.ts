import Chat from "@suy/bots-web-client";
import { Message } from "@suy/bots-web-client/lib/message";
import debug from "debug";

debug.enable("*");

let userMessage: HTMLTemplateElement = null;
let botMessage: HTMLTemplateElement = null;

const log: debug.IDebugger = debug("app");

window.addEventListener("DOMContentLoaded", loaded);

async function loaded(): Promise<void> {
    const chat: Chat = new Chat(`ws://${window.location.host}/echo_chat`);
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

        const pnode: HTMLDivElement = !msg.prev ? null : document.querySelector(`.message[data-ts="${msg.prev.id}"]`);
        const nnode: HTMLDivElement = !msg.next ? null : document.querySelector(`.message[data-ts="${msg.next.id}"]`);

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
    return document.importNode(userMessage.content, true);
}

function newBotMessage(msg: Message): DocumentFragment {
    botMessage.content.querySelector("section").textContent = msg.text;
    botMessage.content.querySelector(".message").setAttribute("data-ts", msg.id.toString());
    return document.importNode(botMessage.content, true);
}