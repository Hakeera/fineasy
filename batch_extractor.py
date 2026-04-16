import os
import csv
import json
from extractor import extrair_pdf

PDF_DIR = "data/pdfs"
CSV_PATH = "data/csv/output.csv"
CONTROLE_PATH = "data/csv/processados.json"  # <-- arquivo de controle


def carregar_processados() -> set:
    """Retorna o conjunto de nomes de arquivos já processados."""
    if not os.path.exists(CONTROLE_PATH):
        return set()
    with open(CONTROLE_PATH, "r", encoding="utf-8") as f:
        return set(json.load(f))


def salvar_processados(processados: set):
    with open(CONTROLE_PATH, "w", encoding="utf-8") as f:
        json.dump(sorted(processados), f, ensure_ascii=False, indent=2)


def main():
    os.makedirs("data/csv", exist_ok=True)

    ja_processados = carregar_processados()

    todos_arquivos = [f for f in os.listdir(PDF_DIR) if f.endswith(".pdf")]
    novos_arquivos = [f for f in todos_arquivos if f not in ja_processados]

    if not novos_arquivos:
        print("Nenhum PDF novo para processar.")
        return

    print(f"{len(novos_arquivos)} novo(s) PDF(s) encontrado(s).")

    novos_resultados = []
    processados_com_sucesso = set()

    for arquivo in novos_arquivos:
        caminho = os.path.join(PDF_DIR, arquivo)
        print(f"Processando: {arquivo}")
        resultado = extrair_pdf(caminho)
        resultado["arquivo"] = arquivo
        novos_resultados.append(resultado)

        # Só marca como processado se não houve erro
        if "error" not in resultado:
            processados_com_sucesso.add(arquivo)
        else:
            print(f"  ⚠ Erro em {arquivo}: {resultado['error']}")

    # Anexa ao CSV existente (modo append)
    csv_existe = os.path.exists(CSV_PATH)
    campos = sorted({k for r in novos_resultados for k in r.keys()})

    with open(CSV_PATH, "a", newline="", encoding="utf-8") as f:
        writer = csv.DictWriter(f, fieldnames=campos)
        if not csv_existe:
            writer.writeheader()
        writer.writerows(novos_resultados)

    # Persiste o controle só após gravar o CSV com sucesso
    salvar_processados(ja_processados | processados_com_sucesso)

    print(f"\nCSV atualizado em: {CSV_PATH}")
    print(f"Processados com sucesso: {len(processados_com_sucesso)}")
    if len(processados_com_sucesso) < len(novos_arquivos):
        erros = len(novos_arquivos) - len(processados_com_sucesso)
        print(f"Com erro (serão tentados novamente): {erros}")


if __name__ == "__main__":
    main()
