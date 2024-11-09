import httpx

import time

def main():
    tries = 0
    success_1 = False
    success_2 = False

    while tries < 10:
        print(f"try: {tries}")
        try:
            response_1 = httpx.get('http://test-1:8000/')
            if response_1.text == "\"example 1\"":
                print("success 1")
                success_1 = True
        except Exception as e:
            print(e)

        try:
            response_2 = httpx.get('http://test-2:8001/')
            if response_2.text == "\"example 2\"":
                print("success 2")
                success_2 = True
        except Exception as e:
            print(e)

        if success_1 is True and success_2 is True:
            print("0")
            exit(0)

        time.sleep(3)
        tries += 1

    print(1)
    exit(1)

main()

