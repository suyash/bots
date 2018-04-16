import debug from "debug";

import Chat from "@suy/bots-web-client";
import {
    Attachment,
    Audio,
    FileDownload,
    Image,
    Location,
    Message,
    Video,
} from "@suy/bots-web-client/lib/message";

debug.enable("*");

let userMessage: HTMLTemplateElement = null;
let botMessage: HTMLTemplateElement = null;

let imageAttachment: HTMLTemplateElement = null;
let audioAttachment: HTMLTemplateElement = null;
let videoAttachment: HTMLTemplateElement = null;
let locationAttachment: HTMLTemplateElement = null;
let fileDownloadAttachment: HTMLTemplateElement = null;

const log: debug.IDebugger = debug("app");

window.addEventListener("DOMContentLoaded", loaded);

async function loaded(): Promise<void> {
    // tslint:disable-next-line:max-line-length
    const chat: Chat = new Chat(`${window.location.protocol === "https:" ? "wss:" : "ws:"}//${window.location.host}/attachments_chat`);
    await chat.open();
    log("open");

    const container: HTMLDivElement = document.querySelector(".messages");

    userMessage = document.querySelector("#user-message");
    botMessage = document.querySelector("#bot-message");

    imageAttachment = document.querySelector("#attachment_image");
    audioAttachment = document.querySelector("#attachment_audio");
    videoAttachment = document.querySelector("#attachment_video");
    locationAttachment = document.querySelector("#attachment_location");
    fileDownloadAttachment = document.querySelector("#attachment_file_download");

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
        msg.text = msg.text.replace("\n", "<br>");

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
    if (msg.text) {
        botMessage.content.querySelector(".text").classList.remove("hidden");
        botMessage.content.querySelector(".text").textContent = msg.text;
    } else {
        botMessage.content.querySelector(".text").classList.add("hidden");
    }

    const attachments: HTMLElement = botMessage.content.querySelector(".attachments");
    while (attachments.firstChild) {
        attachments.removeChild(attachments.firstChild);
    }

    if (msg.attachments && msg.attachments.length) {
        attachments.classList.remove("hidden");

        for (const a of msg.attachments) {
            attachments.appendChild(newAttachment(a));
        }
    } else {
        attachments.classList.add("hidden");
    }

    botMessage.content.querySelector(".message").setAttribute("data-ts", msg.id.toString());
    return document.importNode(botMessage.content, true);
}

function newAttachment(a: Attachment): DocumentFragment {
    switch (a.type) {
    case "image":
        const img: Image = a as Image;
        imageAttachment.content.querySelector("img").src = img.url;
        imageAttachment.content.querySelector("img").alt = img.alt;
        imageAttachment.content.querySelector(".attachment_title").innerHTML = img.title;
        imageAttachment.content.querySelector(".attachment_text").innerHTML = img.text;
        return document.importNode(imageAttachment.content, true);
    case "audio":
        const aud: Audio = a as Audio;
        audioAttachment.content.querySelector("audio").src = aud.url;
        audioAttachment.content.querySelector(".attachment_title").innerHTML = aud.title;
        audioAttachment.content.querySelector(".attachment_text").innerHTML = aud.text;
        return document.importNode(audioAttachment.content, true);
    case "video":
        const vid: Video = a as Video;
        videoAttachment.content.querySelector("video").src = vid.url;
        videoAttachment.content.querySelector(".attachment_title").innerHTML = vid.title;
        videoAttachment.content.querySelector(".attachment_text").innerHTML = vid.text;
        return document.importNode(videoAttachment.content, true);
    case "location":
        const loc: Location = a as Location;
        locationAttachment.content.querySelector(".attachment_title").innerHTML = loc.title;
        locationAttachment.content.querySelector(".attachment_text").innerHTML = loc.text;
        const node: DocumentFragment = document.importNode(locationAttachment.content, true);
        node.querySelector("img").src = googleStaticMapURL(loc);
        return node;
    case "file_download":
        const fd: FileDownload = a as FileDownload;
        fileDownloadAttachment.content.querySelector("a").setAttribute("href", fd.url);
        fileDownloadAttachment.content.querySelector(".attachment_text").innerHTML = fd.text;
        return document.importNode(fileDownloadAttachment.content, true);
    default:
        return null;
    }
}

function googleStaticMapURL(l: Location): string {
    return `https://maps.googleapis.com/maps/api/staticmap?center=${l.lat},${l.long}&zoom=13&size=600x300`;
}
