import fileinput

if __name__ == "__main__":
    for name in fileinput.input():
        print("Hello {}".format(name))

