import {Inter} from "next/font/google";
import {useState} from "react";

const inter = Inter({subsets: ["latin"]});

export default function Home() {
    const [chunkSize, setChunkSize] = useState(20);

    function sendChunk(chunk: Blob, chunkIndex: number, uploadId: string) {
        const formData = new FormData();
        formData.append("chunk", chunk);
        formData.append("chunkIndex", chunkIndex.toString());
        formData.append("uploadId", uploadId);
        return fetch("/api/upload-chunk", {
            method: "POST",
            body: formData
        });
    }

    async function handleFileChange(event: React.ChangeEvent<HTMLInputElement>) {
        if (!event.target.files || event.target.files.length === 0) {
            return; // User canceled file selection
        }

        console.log("Chunk size:", chunkSize);

        const file = event.target.files[0];
        if (file) {
            const fileSizeInMB = file.size / (1024 * 1024);
            const chunkCount = Math.ceil(fileSizeInMB / chunkSize);

            // Split file into chunks
            const chunks: Blob[] = [];
            for (let i = 0; i < chunkCount; i++) {
                const start = i * chunkSize * 1024 * 1024;
                const end = Math.min(start + chunkSize * 1024 * 1024, file.size);
                chunks.push(file.slice(start, end));
            }

            // Request upload with the following parameters: file, chunkCount, chunkSize
            const response = await fetch("/api/request-upload", {
                method: "POST",
                body: JSON.stringify({
                    file,
                    chunkCount,
                    chunkSize
                })
            });

            const data = await response.json();
            if (response.ok) {
                console.log("Upload request succeeded:", data);
            }

            while (chunks.length > 0) {
                // Send remaining chunks
                const response = sendChunk(chunks[0], chunkCount - chunks.length, data.uploadId);
                response.then(value => {
                    if (value.ok) {
                        chunks.shift();
                    } else {
                        console.error("Failed to send chunk:", value);
                    }
                });
            }

            const finalizeResponse = await fetch("/api/finalize-upload", {
                method: "POST",
                body: JSON.stringify({
                    uploadId: data.uploadId
                })
            });

            if (finalizeResponse.ok) {
                console.log("Upload finalized:", data.uploadId);
            }
        }
    }

    return (
        <main className={`flex min-h-screen flex-col space-y-2 items-center p-24 ${inter.className}`}>
            <label htmlFor={"chunk-size"}>
                <p>Chunk size</p>
            </label>
            <input name={"chunk-size"} type="number" placeholder="Chunk size" required
                   onChange={(e) => setChunkSize(e.target.valueAsNumber)}
                   defaultValue={20}/>
            <input type="file" onChange={handleFileChange}/>
        </main>
    );
}
