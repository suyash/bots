import debug from "debug";

import Chat from "@suy/bots-web-client";
import { Message } from "@suy/bots-web-client/lib/message";

debug.enable("*");

const log: debug.IDebugger = debug("app");

let chat: Chat = null;
let userMessage: HTMLTemplateElement = null;
let botMessage: HTMLTemplateElement = null;
let thread: HTMLTemplateElement = null;

window.addEventListener("DOMContentLoaded", loaded);

async function loaded(): Promise<void> {
    chat = new Chat(`${window.location.protocol === "https:" ? "wss:" : "ws:"}//${window.location.host}/threads_chat`);
    await chat.open();
    log("open");

    const container: HTMLDivElement = document.querySelector(".messages");

    userMessage = document.querySelector("#user-message");
    botMessage = document.querySelector("#bot-message");
    thread = document.querySelector("#thread");

    const form: HTMLFormElement = document.querySelector(".form");
    const message: HTMLInputElement = document.querySelector(".form input[type=\"text\"]");
    form.addEventListener("submit", (e: Event): void => {
        e.preventDefault();

        log("sending", message.value);
        chat.say(message.value);

        message.value = "";
    });

    for await (const m of chat.messages()) {
        const msg: Message = m;
        log("received", msg);
        msg.text = msg.text.replace("\n", "<br>");

        let msgContainer: HTMLDivElement = container;
        if (msg.threadId) {
            msgContainer = document.querySelector(`.thread[data-thread-ts="${msg.threadId}"] .messages`); // tslint:disable-line:max-line-length
        }

        let node: DocumentFragment = null;

        if (msg.type === "thread") {
            node = createThread(msg);
        } else if (msg.source === "bot") {
            node = newBotMessage(msg);
        } else {
            node = newUserMessage(msg);
        }

        const pnode: HTMLDivElement = !msg.prev
            ? null
            : msg.prev.type === "message"
                ? document.querySelector(`.message[data-ts="${msg.prev.id}"]`)
                : document.querySelector(`.thread[data-thread-ts="${msg.prev.id}"]`);

        const nnode: HTMLDivElement = !msg.next
            ? null
            : msg.next.type === "message"
                ? document.querySelector(`.message[data-ts="${msg.next.id}"]`)
                : document.querySelector(`.thread[data-thread-ts="${msg.next.id}"]`);

        // TODO: shouldn't need pnode.parentElement === msgContainer
        if (msg.prev && pnode && pnode.parentElement === msgContainer && pnode.nextSibling) {
            pnode.parentElement.insertBefore(node, pnode.nextSibling);
        } else if (msg.next && nnode) {
            nnode.parentElement.insertBefore(node, nnode);
        } else {
            msgContainer.appendChild(node);
        }

        log(document.querySelector(`.message[data-ts="${msg.id}"]`));
    }

    log("closed");
}

function newBotMessage(msg: Message): DocumentFragment {
    botMessage.content.querySelector("section").textContent = msg.text;
    botMessage.content.querySelector(".message").setAttribute("data-ts", msg.id.toString());
    return document.importNode(botMessage.content, true);
}

function newUserMessage(msg: Message): DocumentFragment {
    userMessage.content.querySelector("section").textContent = msg.text;
    userMessage.content.querySelector(".message").setAttribute("data-ts", msg.id.toString());
    return document.importNode(userMessage.content, true);
}

function createThread(msg: Message): DocumentFragment {
    thread.content.querySelector(".thread").setAttribute("data-thread-ts", String(msg.id));

    const fragment: DocumentFragment = document.importNode(thread.content, true);
    const main: HTMLDivElement = fragment.querySelector(".thread");
    const opener: HTMLDivElement = fragment.querySelector(".thread-opener a");
    const form: HTMLFormElement = fragment.querySelector("form");
    const message: HTMLInputElement = fragment.querySelector("form input[type=\"text\"]");

    opener.addEventListener("click", (e: MouseEvent): void => {
        e.preventDefault();
        main.classList.toggle("open");
    });

    form.addEventListener("submit", (e: Event): void => {
        e.preventDefault();
        chat.sayInThread(message.value, msg.id);
        message.value = "";
    });

    return fragment;
}
