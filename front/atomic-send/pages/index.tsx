import {Inter} from "next/font/google";
import {useState} from "react";
import axios from "axios";

const inter = Inter({subsets: ["latin"]});

export default function Home() {
    const [chunkSize, setChunkSize] = useState(20);

    function sendChunk(chunk: Blob, from: number, to: number, total: number, uploadId: string) {
        return axios.post("http://localhost:8080/api/v1/upload-chunk", chunk, {
            headers: {
                "Content-Range": `bytes ${from}-${to} ${total}`,
                "Upload-Id": uploadId,
            }
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
            const startAndEnd: { start: number, end: number }[] = [];
            for (let i = 0; i < chunkCount; i++) {
                const start = i * chunkSize * 1024 * 1024;
                const end = Math.min(start + chunkSize * 1024 * 1024, file.size);
                chunks.push(file.slice(start, end));
                startAndEnd.push({start, end});
            }

            // Request upload with the following parameters: file, chunkCount, chunkSize
            const response = await axios.post("http://localhost:8080/api/v1/request-upload", {
                file_name: file.name,
                chunk_count: chunkCount,
                chunk_size: chunkSize
            });

            const data = await response.data;
            if (response.status === 200) {
                console.log("Upload request succeeded:", data);
            }

            while (chunks.length > 0) {
                // Send remaining chunks
                const response = await sendChunk(chunks[0], startAndEnd[0].start, startAndEnd[0].end, file.size, data.uploadId);
                if (response.status === 200) {
                    chunks.shift();
                    startAndEnd.shift();
                } else {
                    console.error("Failed to send chunk:", response);
                }
            }

            const finalizeResponse = await axios.post("http://localhost:8080/api/v1/finalize-upload", {
                uploadId: data.uploadId
            });

            if (finalizeResponse.status === 200) {
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
