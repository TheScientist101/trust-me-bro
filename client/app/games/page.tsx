"use client";
import { title } from "@/components/primitives";
import { useEffect, useState } from "react";
import { Card } from "@heroui/card";
import {
    Modal,
    ModalContent,
    ModalHeader,
    ModalBody,
    ModalFooter
} from "@heroui/modal";
import { Button } from "@heroui/button";
import { siteConfig } from "@/config/site";

export default function GamesPage() {
    const [accessToken, setAccessToken] = useState<string>("");
    const [accessTokenExpiry, setAccessTokenExpiry] = useState<string>("");
    const [games, setGames] = useState<string[]>([]);
    const [isModalOpen, setIsModalOpen] = useState<boolean>(false);
    const [selectedGame, setSelectedGame] = useState<string | null>(null);

    useEffect(() => {
        setAccessToken(localStorage.getItem("access_token") || "");
        setAccessTokenExpiry(localStorage.getItem("access_token_expiry") || "");
    }, []);

    useEffect(() => {
        if (accessToken == "") {
            return;
        }

        fetch(new URL("/pendingGames", siteConfig.links.server).href, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({
                access_token: accessToken,
            }),
        }).then((res) => res.text())
            .then((text) => JSON.parse(text))
            .then((data) => {
                const playedGames = JSON.parse(localStorage.getItem("played_games") || "[]");
                const filteredGames = data.games.filter((game: string) => !playedGames.includes(game));
                setGames(Array.from(new Set(filteredGames)));
                console.log(data);
            });
    }, [accessToken]);

    const handleCardClick = (game: string) => {
        setSelectedGame(game);
        setIsModalOpen(true);
    };

    const handleModalSubmit = (move: string) => {
        fetch(new URL("/play", siteConfig.links.server).href, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({
                access_token: accessToken,
                game_id: selectedGame,
                move: move,
            }),
        }).then((res) => res.text())
            .then((text) => JSON.parse(text))
            .then((data) => {
                console.log(data);
                const playedGames = JSON.parse(localStorage.getItem("played_games") || "[]");
                playedGames.push(selectedGame);
                localStorage.setItem("played_games", JSON.stringify(playedGames));
                setGames(games.filter((game) => game !== selectedGame)); // Update state immediately
                setIsModalOpen(false);
            });
    };

    return (
        <div className="flex flex-col items-center justify-center gap-4 py-8 md:py-10">
            <h1 className={title()}>Pending Games</h1>
            <ul className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {games.length === 0 ? <p>No matchups available</p> : games.map((game, index) => (
                    <Button key={game} className="p-6 text-center text-lg bg-blue-500 text-white rounded-lg shadow-md hover:bg-blue-700" onPress={() => handleCardClick(game)}>
                        <Card className="p-6 text-center text-lg w-full">
                            <p>{(index + 1) % 3 === 1 ? "Trust" : (index + 1) % 3 === 2 ? "Me" : "Bro"}</p>
                        </Card>
                    </Button>
                ))}
            </ul>
            <Modal isOpen={isModalOpen} onClose={() => setIsModalOpen(false)}>
                <ModalContent>
                    <ModalHeader>
                        <p id="modal-title" aria-setsize={18}>
                            Pick your choice
                        </p>
                    </ModalHeader>
                    <ModalBody className="flex justify-around">
                        <Button className="p-4 bg-red-500 text-white rounded-full hover:bg-red-700" onPress={() => handleModalSubmit("rock")}>✊</Button>
                        <Button className="p-4 bg-green-500 text-white rounded-full hover:bg-green-700" onPress={() => handleModalSubmit("paper")}>📄</Button>
                        <Button className="p-4 bg-yellow-500 text-white rounded-full hover:bg-yellow-700" onPress={() => handleModalSubmit("scissors")}>✌️</Button>
                    </ModalBody>
                    <ModalFooter>
                        <Button className="p-4 bg-gray-500 text-white rounded-lg hover:bg-gray-700" onPress={() => setIsModalOpen(false)}>
                            Close
                        </Button>
                    </ModalFooter>
                </ModalContent>
            </Modal>
        </div >
    );
}