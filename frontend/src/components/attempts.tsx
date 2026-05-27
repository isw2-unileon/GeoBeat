
type Props = {
    attempts: number,
    status: string,
} | null

export function Attempts({ props }: { props: Props }) {

    const attempts = props?.attempts ?? 0;
    const status = props?.status ?? "";

    return (
        <div className='bg-gray-100 rounded-sm absolute top-30 left-15 flex flex-row'>
            <GameSquares attempts={attempts} />
        </div>
    )
}

function GameSquares({ attempts }: { attempts: number }) {
    return (
        <>
            {[...Array(5)].map((_, i) => (
                <div
                    key={i}
                    className={`w-8 h-8 m-2 rounded-sm ${
                        i < attempts ? "bg-yellow-200" : "bg-gray-200"
                    }`}
                />
            ))}
        </>
    );
}